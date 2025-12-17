import type { SupportedLanguages } from "./highlighter";

export enum Status {
	"RUNNING",
	"DONE",
	"ERROR",
	"SKIP",
}

export type InfoType = "helper" | "request" | "response" | "query" | "result";

export interface InfoArg {
	name?: string;
	value: string;
}

export interface TestInfo {
	type: InfoType;
	title: string;
	content: string;
	args?: InfoArg[];
	timestamp: number;
	order: number;
	language?: SupportedLanguages;
}

export interface TestError {
	message: string;
	expected?: unknown;
	actual?: unknown;
}

export class SubTest {
	name: string;
	order: number;
	status: Status = $derived.by(() => {
		if (this.duration === 0) {
			return Status.RUNNING;
		}
		if (this.errors) {
			return Status.ERROR;
		}
		return Status.DONE;
	});
	duration: number = $state(0);
	errors: TestError[] | null = $state(null);
	infos: TestInfo[] = $state([]);

	constructor(name: string, order: number) {
		this.name = name;
		this.order = order;
	}
}

export class File {
	name: string;
	status: Status = $derived.by(() => {
		if (this.subTests.some((subTest) => subTest.status === Status.ERROR)) {
			return Status.ERROR;
		}
		if (this.subTests.some((subTest) => subTest.status === Status.RUNNING)) {
			return Status.RUNNING;
		}
		return Status.DONE;
	});
	subTests: SubTest[] = $state([]);
	infos: TestInfo[] = $state([]);
	duration: number = $state(0);

	constructor(name: string) {
		this.name = name;
	}
}

class Active {
	file: File | undefined = $state();
	test: SubTest | undefined = $state();
}

export const active = new Active();

type EventSubTest = {
	filename: string;
	name: string;
	runner: string;
	errors: TestError[] | null;
	infos: TestInfo[] | null;
	duration: number;
	order?: number;
};

type EventFile = {
	name: string;
	duration: number;
	subTests: EventSubTest[] | null;
	infos: TestInfo[] | null;
};

type InitEvent = {
	type: "init";
	data: Record<string, EventFile>;
};

type StartEvent = {
	type: "start";
	data: EventFile;
};

type StartTestEvent = {
	type: "start_test";
	data: EventSubTest;
};

type EndTestEvent = {
	type: "end_test";
	data: EventSubTest;
};

type EndEvent = {
	type: "end";
	data: EventFile;
};

type InfoEvent = {
	type: "info";
	data: EventSubTest;
};

type FileInfoEvent = {
	type: "file_info";
	data: EventFile;
};

type Event =
	| InitEvent
	| StartEvent
	| StartTestEvent
	| EndTestEvent
	| EndEvent
	| InfoEvent
	| FileInfoEvent;

function createSubTest(subTest: EventSubTest): SubTest {
	const newSubTest = new SubTest(subTest.name, subTest.order ?? 0);
	newSubTest.duration = subTest.duration;
	newSubTest.errors = subTest.errors;
	newSubTest.infos = subTest.infos ?? [];
	return newSubTest;
}

function createFile(file: EventFile): File {
	const newFile = new File(file.name);
	if (file.subTests) {
		newFile.subTests = file.subTests.map(createSubTest);
	}
	newFile.infos = file.infos ?? [];
	newFile.duration = file.duration;
	return newFile;
}

class Watcher {
	files: File[] = $state([]);
	#eventSource?: EventSource;

	constructor() {
		this.#newEventSource();
	}

	destroy() {
		this.#eventSource?.close();
		this.#eventSource = undefined;
	}

	#newEventSource() {
		this.#eventSource?.close();
		this.#eventSource = new EventSource("/events");
		this.#eventSource.onmessage = this.#handleMessage.bind(this);
		this.#eventSource.onerror = this.#handleError.bind(this);
	}

	#handleError() {
		setTimeout(() => {
			this.#newEventSource();
		}, 1000);
	}

	#handleMessage(event: MessageEvent) {
		let data: Event;
		try {
			data = JSON.parse(event.data) as Event;
		} catch {
			console.error("Failed to parse event data:", event.data);
			return;
		}
		switch (data.type) {
			case "init":
				this.files = Object.values(data.data)
					.map(createFile)
					.sort((a, b) => a.name.localeCompare(b.name));
				break;
			case "start":
			case "end": {
				const existing = this.files.find((file) => file.name === data.data.name);
				if (existing) {
					existing.duration = data.data.duration;
					existing.infos = data.data.infos ?? [];
				} else {
					const file = createFile(data.data);
					this.files.push(file);
				}

				break;
			}
			case "file_info": {
				const existing = this.files.find((file) => file.name === data.data.name);
				if (existing) {
					existing.infos = data.data.infos ?? [];
				}
				break;
			}
			case "info":
			case "end_test":
			case "start_test": {
				const file = this.files.find((file) => file.name === data.data.filename);

				if (!file) {
					return;
				}

				const existingSubTest = file.subTests.find((subTest) => subTest.name === data.data.name);

				if (existingSubTest) {
					existingSubTest.duration = data.data.duration;
					existingSubTest.errors = data.data.errors;
					existingSubTest.infos = data.data.infos ?? [];
				} else {
					const subTest = createSubTest(data.data);
					file.subTests.push(subTest);
				}
				break;
			}
		}
	}
}

export const watcher = new Watcher();
