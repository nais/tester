<script lang="ts">
	import FileButton from "./lib/FileButton.svelte";
	import MessageFormatter from "./lib/MessageFormatter.svelte";
	import { active, watcher } from "./lib/watcher.svelte";

	let fileFilter = $state("");
	let testFilter = $state("");

	const files = $derived.by(() => {
		return watcher.files
			.filter((f) => (fileFilter ? f.name.includes(fileFilter) : true))
			.sort((a, b) => (a.status == b.status ? a.name.localeCompare(b.name) : b.status - a.status));
	});

	const tests = $derived.by(() => {
		return active.file?.subTests
			.filter((f) => (testFilter ? f.name.includes(testFilter) : true))
			.sort((a, b) => a.name.localeCompare(b.name));
	});
</script>

<div id="wrapper">
	<section id="sidebar">
		<input type="search" placeholder="Search files" bind:value={fileFilter} />

		{#each files as file}
			<FileButton
				{file}
				onclick={(name) => {
					active.file = watcher.files.find((f) => f.name == name);
				}}
				active={active.file?.name == file.name}
			/>
		{/each}
	</section>

	<section id="details">
		<input type="search" placeholder="Search" bind:value={testFilter} />
		{#if !tests}
			<p>Select a file</p>
		{:else}
			{#each tests as subTest}
				<FileButton
					file={subTest}
					onclick={(name) => {
						active.test = active.file?.subTests.find((f) => f.name == name);
					}}
					active={active.test?.name == subTest.name}
				/>
			{/each}
		{/if}
	</section>

	<main>
		{#if active.test}
			{#each active.test.errors || [] as { message }}
				<MessageFormatter {message} />
			{:else}
				<p>No errors</p>
			{/each}
		{:else}
			<p>Select a test</p>
		{/if}
	</main>
</div>

<style>
	#wrapper {
		display: grid;
		grid-template-columns: 300px 300px 1fr;
		gap: 1rem;
		padding: 0.5rem;
	}

	#sidebar {
		display: flex;
		flex-direction: column;
		gap: 0.1rem;
	}

	input {
		width: 100%;
	}
</style>
