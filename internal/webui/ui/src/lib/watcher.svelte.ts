export enum Status {
  "RUNNING",
  "DONE",
  "ERROR",
  "SKIP",
}

export class SubTest {
  name: string;
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
  errors: { message: string }[] | null = $state(null);

  constructor(name: string) {
    this.name = name;
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
  errors: { message: string }[] | null;
  duration: number;
};

type EventFile = {
  name: string;
  duration: number;
  subTests: EventSubTest[] | null;
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

type ErrorEvent = {
  type: "error";
  data: EventSubTest;
};

type EndEvent = {
  type: "end";
  data: EventFile;
};

type Event = InitEvent | StartEvent | StartTestEvent | EndTestEvent | EndEvent;

function createSubTest(subTest: EventSubTest): SubTest {
  const newSubTest = new SubTest(subTest.name);
  newSubTest.duration = subTest.duration;
  newSubTest.errors = subTest.errors;
  return newSubTest;
}

function createFile(file: EventFile): File {
  const newFile = new File(file.name);
  // newFile.status = $state(Status.DONE);
  if (file.subTests) {
    console.log("HAS", file.subTests);
    newFile.subTests = file.subTests.map(createSubTest);
  }
  newFile.duration = file.duration;
  return newFile;
}

class Watcher {
  files: File[] = $state([]);

  constructor() {
    new EventSource("/events").onmessage = (event) => {
      const data = JSON.parse(event.data) as Event;
      switch (data.type) {
        case "init":
          this.files = Object.values(data.data)
            .map(createFile)
            .sort((a, b) => a.name.localeCompare(b.name));
          break;
        case "start":
        case "end":
          const existing = this.files.find(
            (file) => file.name === data.data.name
          );
          if (existing) {
            existing.duration = data.data.duration;
            existing.subTests = data.data.subTests?.map(createSubTest) || [];
          } else {
            const file = createFile(data.data);
            this.files.push(file);
          }

          break;

        // Ignore these events for now
        // case "error":
        //   break;
        // case "start_test":
        // case "end_test":
        //   const file = this.files.find(
        //     (file) => file.name === data.data.filename
        //   );
        //   if (file) {
        //     const existing = file.subTests.find(
        //       (subTest) => subTest.name === data.data.name
        //     );

        //     if (existing) {
        //       existing.duration = data.data.duration;
        //       existing.errors = data.data.errors;
        //     } else {
        //       const subTest = createSubTest(data.data);
        //       file.subTests.push(subTest);
        //     }
        //   }
        //   break;
      }
    };
  }
}

export const watcher = new Watcher();
