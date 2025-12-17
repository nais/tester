<script lang="ts">
	import { fuzzySearch } from "./fuzzy";
	import { Status, watcher, type File, type SubTest } from "./watcher.svelte";

	interface Props {
		isOpen: boolean;
		onClose: () => void;
		onSelectFile: (file: File) => void;
		onSelectTest: (test: SubTest, file: File) => void;
	}

	let { isOpen = $bindable(), onClose, onSelectFile, onSelectTest }: Props = $props();

	let dialogElement: HTMLDialogElement | undefined = $state();
	let searchInput: HTMLInputElement | undefined = $state();
	let query = $state("");
	let selectedIndex = $state(0);

	// Determine search mode based on query prefix
	const searchMode = $derived.by(() => {
		if (query.startsWith(":")) {
			return "files";
		}
		if (query.startsWith("@")) {
			return "tests";
		}
		return "all";
	});

	// Get the actual search pattern (remove prefix if present)
	const searchPattern = $derived.by(() => {
		if (query.startsWith(":") || query.startsWith("@")) {
			return query.slice(1);
		}
		return query;
	});

	// File results
	const fileResults = $derived.by(() => {
		if (searchMode === "tests") {
			return [];
		}
		return fuzzySearch(watcher.files, searchPattern, (file) => file.name);
	});

	// Test results (flatten all tests from all files)
	const testResults = $derived.by(() => {
		if (searchMode === "files") {
			return [];
		}

		const allTests: Array<{ test: SubTest; file: File }> = [];
		for (const file of watcher.files) {
			for (const test of file.subTests) {
				allTests.push({ test, file });
			}
		}

		return fuzzySearch(allTests, searchPattern, (item) => item.test.name);
	});

	// Combined results
	const allResults = $derived.by(() => {
		const results: Array<{ type: "file" | "test"; item: File | SubTest; file?: File }> = [];

		// Add files
		for (const file of fileResults) {
			results.push({ type: "file", item: file });
		}

		// Add tests
		for (const { test, file } of testResults) {
			results.push({ type: "test", item: test, file });
		}

		// Limit results to prevent performance issues
		return results.slice(0, 50);
	});

	// Reset state when dialog opens/closes
	$effect(() => {
		if (isOpen && dialogElement) {
			dialogElement.showModal();
			query = "";
			selectedIndex = 0;
			// Focus input after a tick to ensure dialog is fully rendered
			setTimeout(() => searchInput?.focus(), 0);
		} else if (!isOpen && dialogElement) {
			dialogElement.close();
		}
	});

	// Reset selected index when results change
	$effect(() => {
		if (allResults.length > 0) {
			selectedIndex = Math.min(selectedIndex, allResults.length - 1);
		} else {
			selectedIndex = 0;
		}
	});

	function handleClose() {
		isOpen = false;
		onClose();
	}

	function handleKeydown(e: KeyboardEvent) {
		switch (e.key) {
			case "Escape":
				e.preventDefault();
				handleClose();
				break;
			case "ArrowDown":
				e.preventDefault();
				selectedIndex = Math.min(selectedIndex + 1, allResults.length - 1);
				scrollToSelected();
				break;
			case "ArrowUp":
				e.preventDefault();
				selectedIndex = Math.max(selectedIndex - 1, 0);
				scrollToSelected();
				break;
			case "Enter":
				e.preventDefault();
				selectCurrent();
				break;
		}
	}

	function scrollToSelected() {
		const selected = dialogElement?.querySelector(`[data-index="${selectedIndex}"]`);
		if (selected) {
			selected.scrollIntoView({ block: "nearest", behavior: "smooth" });
		}
	}

	function selectCurrent() {
		const result = allResults[selectedIndex];
		if (!result) return;

		if (result.type === "file") {
			onSelectFile(result.item as File);
		} else if (result.type === "test" && result.file) {
			onSelectTest(result.item as SubTest, result.file);
		}

		handleClose();
	}

	function selectItem(index: number) {
		selectedIndex = index;
		selectCurrent();
	}

	function getStatusIcon(status: Status): string {
		switch (status) {
			case Status.ERROR:
				return "✕";
			case Status.DONE:
				return "✓";
			case Status.RUNNING:
				return "●";
			default:
				return "○";
		}
	}

	function getStatusClass(status: Status): string {
		switch (status) {
			case Status.ERROR:
				return "error";
			case Status.DONE:
				return "success";
			case Status.RUNNING:
				return "running";
			default:
				return "";
		}
	}
</script>

<dialog
	bind:this={dialogElement}
	class="command-panel"
	onclose={handleClose}
	onkeydown={handleKeydown}
	onclick={(e) => {
		// Close when clicking on backdrop (outside the dialog content)
		if (e.target === dialogElement) {
			handleClose();
		}
	}}
>
	<div class="panel-content">
		<div class="search-box">
			<div class="search-icon">
				{#if searchMode === "files"}
					Files
				{:else if searchMode === "tests"}
					Tests
				{:else}
					⌘
				{/if}
			</div>
			<input
				bind:this={searchInput}
				bind:value={query}
				type="text"
				placeholder="Search files and tests... (prefix with : for files only, @ for tests only)"
				class="search-input"
				autocomplete="off"
				spellcheck="false"
			/>
			{#if query}
				<button class="clear-button" onclick={() => (query = "")} type="button">✕</button>
			{/if}
		</div>

		{#if query && allResults.length === 0}
			<div class="no-results">
				<p>No results found for "{searchPattern}"</p>
				{#if searchMode !== "all"}
					<p class="hint">Try removing the prefix to search everything</p>
				{/if}
			</div>
		{:else}
			<div class="results">
				{#if !query}
					<div class="hints">
						<div class="hint-item">
							<kbd>:</kbd>
							<span>Search files only</span>
						</div>
						<div class="hint-item">
							<kbd>@</kbd>
							<span>Search tests only</span>
						</div>
						<div class="hint-item">
							<kbd>↑↓</kbd>
							<span>Navigate</span>
						</div>
						<div class="hint-item">
							<kbd>Enter</kbd>
							<span>Select</span>
						</div>
						<div class="hint-item">
							<kbd>Esc</kbd>
							<span>Close</span>
						</div>
					</div>
				{:else}
					{#each allResults as result, index (result.type === "file" ? "file-" + (result.item as File).name : "test-" + result.file!.name + "-" + (result.item as SubTest).name)}
						<button
							class="result-item"
							class:selected={index === selectedIndex}
							data-index={index}
							onclick={() => selectItem(index)}
							type="button"
						>
							{#if result.type === "file"}
								{@const file = result.item as File}
								<span class="status-icon {getStatusClass(file.status)}">
									{getStatusIcon(file.status)}
								</span>
								<span class="result-name">{file.name}</span>
								<span class="result-badge file-badge">File</span>
								<span class="result-meta">
									{file.subTests.length}
									{file.subTests.length === 1 ? "test" : "tests"}
								</span>
							{:else}
								{@const test = result.item as SubTest}
								{@const file = result.file!}
								<span class="status-icon {getStatusClass(test.status)}">
									{getStatusIcon(test.status)}
								</span>
								<span class="result-name">{test.name}</span>
								<span class="result-badge test-badge">Test</span>
								<span class="result-meta">{file.name}</span>
							{/if}
						</button>
					{/each}
				{/if}
			</div>
		{/if}

		<div class="footer">
			<div class="footer-hint">
				{#if query && allResults.length > 0}
					{allResults.length} result{allResults.length === 1 ? "" : "s"}
				{:else if !query}
					{watcher.files.length} files, {watcher.files.reduce(
						(acc, f) => acc + f.subTests.length,
						0,
					)} tests
				{/if}
			</div>
		</div>
	</div>
</dialog>

<style>
	.command-panel {
		border: none;
		border-radius: var(--radius-md, 8px);
		padding: 0;
		max-width: 640px;
		width: 90vw;
		max-height: 60vh;
		background: var(--color-bg-elevated);
		box-shadow:
			0 20px 25px -5px rgb(0 0 0 / 0.1),
			0 8px 10px -6px rgb(0 0 0 / 0.1);
		overflow: hidden;
		color: var(--color-text);
		position: fixed;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		margin: 0;
	}

	.command-panel::backdrop {
		background: rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(4px);
	}

	.panel-content {
		display: flex;
		flex-direction: column;
		height: 100%;
	}

	.search-box {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 1rem;
		border-bottom: 1px solid var(--color-border);
		background: var(--color-bg-elevated);
	}

	.search-icon {
		font-size: 1.25rem;
		color: var(--color-text-muted);
		width: 2.5rem;
		height: 1.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.search-input {
		flex: 1;
		background: transparent;
		border: none;
		outline: none;
		font-size: 1rem;
		color: var(--color-text);
		padding: 0;
	}

	.search-input::placeholder {
		color: var(--color-text-muted);
	}

	.clear-button {
		background: transparent;
		border: none;
		color: var(--color-text-muted);
		cursor: pointer;
		padding: 0.25rem;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: var(--radius-sm, 4px);
		width: 1.5rem;
		height: 1.5rem;
		flex-shrink: 0;
	}

	.clear-button:hover {
		background: var(--color-bg-hover);
		color: var(--color-text);
	}

	.results {
		flex: 1;
		overflow-y: auto;
		min-height: 200px;
		max-height: calc(60vh - 120px);
	}

	.hints {
		padding: 2rem;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.hint-item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		font-size: 0.875rem;
		color: var(--color-text-muted);
	}

	kbd {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 1.5rem;
		height: 1.5rem;
		padding: 0 0.375rem;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm, 4px);
		font-family: ui-monospace, monospace;
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--color-text);
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
	}

	.result-item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem 1rem;
		width: 100%;
		background: transparent;
		border: none;
		border-bottom: 1px solid var(--color-border);
		cursor: pointer;
		text-align: left;
		color: var(--color-text);
		transition: background-color 0.1s ease;
	}

	.result-item:hover,
	.result-item.selected {
		background: var(--color-bg-active);
	}

	.result-item.selected {
		outline: 2px solid var(--color-running);
		outline-offset: -2px;
	}

	.status-icon {
		width: 1.25rem;
		height: 1.25rem;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.75rem;
		font-weight: bold;
		flex-shrink: 0;
	}

	.status-icon.success {
		color: var(--color-success);
	}

	.status-icon.error {
		color: var(--color-error);
	}

	.status-icon.running {
		color: var(--color-running);
		animation: pulse 2s infinite;
	}

	.result-name {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-size: 0.9375rem;
	}

	.result-badge {
		font-size: 0.625rem;
		padding: 0.25rem 0.5rem;
		border-radius: 9999px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		flex-shrink: 0;
	}

	.file-badge {
		background: color-mix(in srgb, var(--color-running) 20%, transparent);
		color: var(--color-running);
	}

	.test-badge {
		background: color-mix(in srgb, var(--color-text-muted) 20%, transparent);
		color: var(--color-text-muted);
	}

	.result-meta {
		font-size: 0.75rem;
		color: var(--color-text-muted);
		flex-shrink: 0;
	}

	.no-results {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem 2rem;
		text-align: center;
		color: var(--color-text-muted);
	}

	.no-results p {
		margin: 0;
	}

	.no-results .hint {
		font-size: 0.875rem;
		margin-top: 0.5rem;
	}

	.footer {
		padding: 0.5rem 1rem;
		border-top: 1px solid var(--color-border);
		background: var(--color-bg);
	}

	.footer-hint {
		font-size: 0.75rem;
		color: var(--color-text-muted);
		text-align: center;
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
