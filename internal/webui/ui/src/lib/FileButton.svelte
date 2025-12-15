<script lang="ts">
	import { formatNanoseconds } from "./format";
	import Check from "./icons/Check.svelte";
	import Error from "./icons/Error.svelte";
	import Record from "./icons/Record.svelte";
	import Skip from "./icons/Skip.svelte";
	import { Status } from "./watcher.svelte";
	let {
		file,
		active,
		onclick,
	}: {
		file: { name: string; status: Status; duration: number };
		active: boolean;
		onclick: (name: string) => void;
	} = $props();
</script>

<button onclick={() => onclick(file.name)} class:active>
	<span
		class:error={file.status == Status.ERROR}
		class:done={file.status == Status.DONE}
		class:running={file.status == Status.RUNNING}
	>
		{#if file.status == Status.RUNNING}
			<Record />
		{:else if file.status == Status.ERROR}
			<Error />
		{:else if file.status == Status.SKIP}
			<Skip />
		{:else}
			<Check />
		{/if}
	</span>
	{file.name}
	<span class="duration">
		{#if file.status == Status.RUNNING}
			running...
		{:else}
			{formatNanoseconds(file.duration)}
		{/if}
	</span>
</button>

<style>
	button {
		font-size: 0.9rem;
		background-color: transparent;
		border: none;
		text-align: left;
		display: flex;
		justify-content: flex-start;
		align-items: center;
		gap: 0.5rem;

		> * {
			flex-shrink: 1;
		}

		&:hover {
			text-decoration: underline;
			cursor: pointer;
		}

		&.active {
			background-color: #670f8a;
		}
	}

	.error {
		color: rgb(209, 44, 44);
	}

	.done {
		color: rgb(10, 243, 10);
	}

	.duration {
		margin-left: auto;
		opacity: 0.5;
	}

	.running {
		color: rgb(255, 255, 255);
		animation: pulse 2s infinite;
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
