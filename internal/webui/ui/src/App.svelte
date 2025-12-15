<script lang="ts">
	import FileButton from "./lib/FileButton.svelte";
	import MessageFormatter from "./lib/MessageFormatter.svelte";
	import { active, watcher } from "./lib/watcher.svelte";

	let fileFilter = $state("");
	let testFilter = $state("");

	const files = $derived(
		watcher.files
			.filter((f) => (fileFilter ? f.name.includes(fileFilter) : true))
			.sort((a, b) => (a.status === b.status ? a.name.localeCompare(b.name) : b.status - a.status)),
	);

	const tests = $derived(
		active.file?.subTests
			.filter((f) => (testFilter ? f.name.includes(testFilter) : true))
			.sort((a, b) => a.name.localeCompare(b.name)),
	);
</script>

<div id="wrapper">
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
	</aside>

	<main class="panel">
		<header>
			<h2>Output</h2>
			{#if active.test}
				<span class="test-name">{active.test.name}</span>
			{/if}
		</header>
		<div class="content">
			{#if !active.test}
				<p class="empty">Select a test to view output</p>
			{:else if active.test.errors && active.test.errors.length > 0}
				{#each active.test.errors as { message }, i (i)}
					<MessageFormatter {message} />
				{/each}
			{:else}
				<div class="success">
					<span class="success-icon">✓</span>
					<p>Test passed with no errors</p>
				</div>
			{/if}
		</div>
	</main>
</div>

<style>
	#wrapper {
		display: grid;
		grid-template-columns: 280px 280px 1fr;
		height: 100vh;
		overflow: hidden;
	}

	.panel {
		display: flex;
		flex-direction: column;
		border-right: 1px solid var(--color-border);
		overflow: hidden;
	}

	.panel:last-child {
		border-right: none;
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
</style>
