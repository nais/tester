<script lang="ts">
	import FileButton from "./lib/FileButton.svelte";
	import { formatNanoseconds } from "./lib/format";
	import InfoCard from "./lib/InfoCard.svelte";
	import InfoRow from "./lib/InfoRow.svelte";
	import MessageFormatter from "./lib/MessageFormatter.svelte";
	import { active, Status, watcher, type SubTest, type TestInfo } from "./lib/watcher.svelte";

	const STORAGE_KEY = "tester-panel-widths";
	const DEFAULT_FILES_WIDTH = 280;
	const DEFAULT_TESTS_WIDTH = 280;
	const MIN_WIDTH = 150;
	const MAX_WIDTH = 500;

	let fileFilter = $state("");
	let testFilter = $state("");

	// Load saved widths from localStorage
	function loadWidths(): { filesWidth: number; testsWidth: number } {
		try {
			const saved = localStorage.getItem(STORAGE_KEY);
			if (saved) {
				const parsed = JSON.parse(saved);
				return {
					filesWidth: Math.max(
						MIN_WIDTH,
						Math.min(MAX_WIDTH, parsed.filesWidth ?? DEFAULT_FILES_WIDTH),
					),
					testsWidth: Math.max(
						MIN_WIDTH,
						Math.min(MAX_WIDTH, parsed.testsWidth ?? DEFAULT_TESTS_WIDTH),
					),
				};
			}
		} catch {
			// Ignore parse errors
		}
		return { filesWidth: DEFAULT_FILES_WIDTH, testsWidth: DEFAULT_TESTS_WIDTH };
	}

	function saveWidths(filesWidth: number, testsWidth: number) {
		try {
			localStorage.setItem(STORAGE_KEY, JSON.stringify({ filesWidth, testsWidth }));
		} catch {
			// Ignore storage errors
		}
	}

	const initialWidths = loadWidths();
	let filesWidth = $state(initialWidths.filesWidth);
	let testsWidth = $state(initialWidths.testsWidth);

	let dragging: "files" | "tests" | null = $state(null);

	function handleMouseDown(panel: "files" | "tests") {
		return (e: MouseEvent) => {
			e.preventDefault();
			dragging = panel;
		};
	}

	function handleMouseMove(e: MouseEvent) {
		if (!dragging) return;

		if (dragging === "files") {
			const newWidth = Math.max(MIN_WIDTH, Math.min(MAX_WIDTH, e.clientX));
			filesWidth = newWidth;
		} else if (dragging === "tests") {
			const newWidth = Math.max(MIN_WIDTH, Math.min(MAX_WIDTH, e.clientX - filesWidth));
			testsWidth = newWidth;
		}
	}

	function handleMouseUp() {
		if (dragging) {
			saveWidths(filesWidth, testsWidth);
			dragging = null;
		}
	}

	const files = $derived(
		watcher.files
			.filter((f) => (fileFilter ? f.name.includes(fileFilter) : true))
			.sort((a, b) => (a.status === b.status ? a.name.localeCompare(b.name) : b.status - a.status)),
	);

	const tests = $derived(
		active.file?.subTests
			.filter((f) => (testFilter ? f.name.includes(testFilter) : true))
			.sort((a, b) => a.order - b.order),
	);

	// Combined file items (infos + tests) sorted by execution order
	type FileItem = { kind: "info"; data: TestInfo } | { kind: "test"; data: SubTest };

	const fileItems = $derived.by((): FileItem[] => {
		if (!active.file) return [];
		const infos: FileItem[] = (active.file.infos ?? []).map((info) => ({
			kind: "info",
			data: info,
		}));
		const tests: FileItem[] = (active.file.subTests ?? []).map((test) => ({
			kind: "test",
			data: test,
		}));
		return [...infos, ...tests].sort((a, b) => a.data.order - b.data.order);
	});

	// File summary stats
	const fileSummary = $derived.by(() => {
		if (!active.file) return null;
		const allTests = active.file.subTests;
		const passed = allTests.filter((t) => t.status === Status.DONE).length;
		const failed = allTests.filter((t) => t.status === Status.ERROR).length;
		const running = allTests.filter((t) => t.status === Status.RUNNING).length;
		return { total: allTests.length, passed, failed, running };
	});
</script>

<svelte:window onmousemove={handleMouseMove} onmouseup={handleMouseUp} />

<div id="wrapper" style:--files-width="{filesWidth}px" style:--tests-width="{testsWidth}px">
	<aside class="panel">
		<header>
			<h2>Files</h2>
			<span class="count">{files.length}</span>
		</header>
		<input type="search" placeholder="Filter files..." bind:value={fileFilter} />
		<div class="list">
			{#each files as file (file.name)}
				<FileButton
					{file}
					onselect={(name) => {
						active.file = watcher.files.find((f) => f.name === name);
						active.test = undefined;
					}}
					active={active.file?.name === file.name}
				/>
			{:else}
				<p class="empty">No files found</p>
			{/each}
		</div>
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<div
			class="resize-handle"
			class:active={dragging === "files"}
			onmousedown={handleMouseDown("files")}
			role="separator"
			aria-orientation="vertical"
			aria-label="Resize files panel"
		></div>
	</aside>

	<aside class="panel">
		<header>
			<h2>Tests</h2>
			{#if tests}
				<span class="count">{tests.length}</span>
			{/if}
		</header>
		<input
			type="search"
			placeholder="Filter tests..."
			bind:value={testFilter}
			disabled={!active.file}
		/>
		<div class="list">
			{#if !active.file}
				<p class="empty">Select a file to view tests</p>
			{:else if tests}
				{#each tests as subTest (subTest.name)}
					<FileButton
						file={subTest}
						onselect={(name) => {
							active.test = active.file?.subTests.find((f) => f.name === name);
						}}
						active={active.test?.name === subTest.name}
					/>
				{:else}
					<p class="empty">No tests match filter</p>
				{/each}
			{/if}
		</div>
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<div
			class="resize-handle"
			class:active={dragging === "tests"}
			onmousedown={handleMouseDown("tests")}
			role="separator"
			aria-orientation="vertical"
			aria-label="Resize tests panel"
		></div>
	</aside>

	<main class="panel">
		<header>
			<h2>Output</h2>
			{#if active.test}
				<span class="test-name">{active.test.name}</span>
			{:else if active.file}
				<span class="test-name">{active.file.name}</span>
			{/if}
		</header>
		<div class="content">
			{#if !active.file}
				<p class="empty">Select a file to view tests</p>
			{:else if !active.test}
				<!-- File Summary View -->
				<div class="file-summary">
					<div class="summary-header">
						<h3>{active.file.name}</h3>
						<span class="duration">{formatNanoseconds(active.file.duration)}</span>
					</div>

					{#if fileSummary}
						<div class="summary-stats">
							<div class="stat">
								<span class="stat-value">{fileSummary.total}</span>
								<span class="stat-label">Total</span>
							</div>
							<div class="stat stat-passed">
								<span class="stat-value">{fileSummary.passed}</span>
								<span class="stat-label">Passed</span>
							</div>
							<div class="stat stat-failed">
								<span class="stat-value">{fileSummary.failed}</span>
								<span class="stat-label">Failed</span>
							</div>
							{#if fileSummary.running > 0}
								<div class="stat stat-running">
									<span class="stat-value">{fileSummary.running}</span>
									<span class="stat-label">Running</span>
								</div>
							{/if}
						</div>
					{/if}

					<div class="test-list">
						{#each fileItems as item (item.kind + "-" + item.data.order)}
							{#if item.kind === "info"}
								<InfoRow info={item.data} />
							{:else}
								<button
									class="test-row"
									class:error={item.data.status === Status.ERROR}
									class:success={item.data.status === Status.DONE}
									class:running={item.data.status === Status.RUNNING}
									onclick={() => (active.test = item.data)}
								>
									<span class="status-icon">
										{#if item.data.status === Status.ERROR}✕{:else if item.data.status === Status.DONE}✓{:else if item.data.status === Status.RUNNING}●{:else}○{/if}
									</span>
									<span class="test-name">{item.data.name}</span>
									{#if item.data.errors && item.data.errors.length > 0}
										<span class="error-badge">{item.data.errors.length}</span>
									{/if}
									<span class="test-duration">
										{#if item.data.status === Status.RUNNING}running...{:else}{formatNanoseconds(
												item.data.duration,
											)}{/if}
									</span>
								</button>
							{/if}
						{/each}
					</div>
				</div>
			{:else}
				{#if active.test.errors && active.test.errors.length > 0}
					<section class="output-section">
						<h3 class="section-title error-title">Errors</h3>
						{#each active.test.errors as error, i (i)}
							<MessageFormatter {error} />
						{/each}
					</section>
				{:else}
					<div class="success">
						<span class="success-icon">✓</span>
						<p>Test passed with no errors</p>
					</div>
				{/if}

				{#if active.test.infos && active.test.infos.length > 0}
					<section class="output-section">
						<h3 class="section-title">Execution Log</h3>
						{#each active.test.infos as info, i (i)}
							<InfoCard {info} />
						{/each}
					</section>
				{/if}
			{/if}
		</div>
	</main>
</div>

{#if dragging}
	<div class="drag-overlay"></div>
{/if}

<style>
	#wrapper {
		display: grid;
		grid-template-columns: var(--files-width) var(--tests-width) 1fr;
		height: 100vh;
		overflow: hidden;
	}

	.panel {
		display: flex;
		flex-direction: column;
		border-right: 1px solid var(--color-border);
		overflow: hidden;
		position: relative;
	}

	.panel:last-child {
		border-right: none;
	}

	.resize-handle {
		position: absolute;
		top: 0;
		right: 0;
		width: 4px;
		height: 100%;
		cursor: col-resize;
		background: transparent;
		transition: background-color 0.15s ease;
		z-index: 10;
	}

	.resize-handle:hover,
	.resize-handle.active {
		background: var(--color-running);
	}

	.drag-overlay {
		position: fixed;
		inset: 0;
		cursor: col-resize;
		z-index: 100;
	}

	header {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1rem;
		background: var(--color-bg-elevated);
		border-bottom: 1px solid var(--color-border);
		flex-shrink: 0;
	}

	h2 {
		font-size: 0.875rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--color-text-muted);
	}

	.count {
		font-size: 0.75rem;
		padding: 0.125rem 0.5rem;
		background: var(--color-bg-active);
		border-radius: 9999px;
		color: var(--color-text-muted);
	}

	.test-name {
		font-size: 0.75rem;
		color: var(--color-text-muted);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	input[type="search"] {
		margin: 0.5rem;
		padding: 0.5rem 0.75rem;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		color: var(--color-text);
		font-size: 0.875rem;
		flex-shrink: 0;
	}

	input[type="search"]::placeholder {
		color: var(--color-text-muted);
	}

	input[type="search"]:focus {
		outline: none;
		border-color: var(--color-running);
	}

	input[type="search"]:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.list {
		flex: 1;
		overflow-y: auto;
		padding: 0.25rem;
	}

	.content {
		flex: 1;
		overflow-y: auto;
		padding: 1rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.empty {
		padding: 2rem 1rem;
		text-align: center;
		color: var(--color-text-muted);
		font-size: 0.875rem;
	}

	.success {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		padding: 3rem;
		color: var(--color-success);
	}

	.success-icon {
		font-size: 2rem;
		width: 3rem;
		height: 3rem;
		display: flex;
		align-items: center;
		justify-content: center;
		background: color-mix(in srgb, var(--color-success) 15%, transparent);
		border-radius: 50%;
	}

	.success p {
		color: var(--color-text-muted);
	}

	.output-section {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.section-title {
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--color-text-muted);
		margin-bottom: 0.25rem;
	}

	.error-title {
		color: var(--color-error);
	}

	/* File Summary Styles */
	.file-summary {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.summary-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
	}

	.summary-header h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--color-text);
	}

	.summary-header .duration {
		font-size: 0.875rem;
		color: var(--color-text-muted);
	}

	.summary-stats {
		display: flex;
		gap: 1rem;
	}

	.stat {
		flex: 1;
		padding: 1rem;
		background: var(--color-bg-elevated);
		border-radius: var(--radius-sm);
		border: 1px solid var(--color-border);
		text-align: center;
	}

	.stat-value {
		display: block;
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--color-text);
	}

	.stat-label {
		display: block;
		font-size: 0.75rem;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.stat-passed .stat-value {
		color: var(--color-success);
	}

	.stat-failed .stat-value {
		color: var(--color-error);
	}

	.stat-running .stat-value {
		color: var(--color-running);
	}

	.test-list {
		display: flex;
		flex-direction: column;
	}

	.test-row {
		display: flex;
		flex-direction: row;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.75rem;
		background: none;
		border: none;
		border-bottom: 1px solid var(--color-border);
		cursor: pointer;
		text-align: left;
		font-size: 0.8125rem;
		color: var(--color-text);
		width: 100%;
	}

	.test-row:hover {
		background: var(--color-bg-hover);
	}

	.test-row .status-icon {
		width: 1rem;
		font-weight: bold;
		font-size: 0.75rem;
		flex-shrink: 0;
	}

	.test-row.success .status-icon {
		color: var(--color-success);
	}

	.test-row.error .status-icon {
		color: var(--color-error);
	}

	.test-row.running .status-icon {
		color: var(--color-running);
		animation: pulse 2s infinite;
	}

	.test-row .test-name {
		flex: 1;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.test-row .error-badge {
		font-size: 0.625rem;
		padding: 0.125rem 0.375rem;
		background: var(--color-error);
		color: var(--color-bg);
		border-radius: 3px;
		font-weight: 600;
		flex-shrink: 0;
	}

	.test-row .test-duration {
		color: var(--color-text-muted);
		font-size: 0.75rem;
		flex-shrink: 0;
	}

	@keyframes pulse {
		0% {
			opacity: 0.5;
		}
		50% {
			opacity: 1;
		}
		100% {
			opacity: 0.5;
		}
	}
</style>
