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
		onselect,
	}: {
		file: { name: string; status: Status; duration: number };
		active: boolean;
		onselect: (name: string) => void;
	} = $props();
</script>

<button onclick={() => onselect(file.name)} class:active>
	<span
		class="icon"
		class:error={file.status === Status.ERROR}
		class:done={file.status === Status.DONE}
		class:running={file.status === Status.RUNNING}
		class:skip={file.status === Status.SKIP}
	>
		{#if file.status === Status.RUNNING}
			<Record />
		{:else if file.status === Status.ERROR}
			<Error />
		{:else if file.status === Status.SKIP}
			<Skip />
		{:else}
			<Check />
		{/if}
	</span>
	<span class="name">{file.name}</span>
	<span class="duration">
		{#if file.status === Status.RUNNING}
			running
		{:else}
			{formatNanoseconds(file.duration)}
		{/if}
	</span>
</button>

<style>
	button {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		width: 100%;
		padding: 0.5rem 0.75rem;
		background: transparent;
		border: none;
		border-radius: var(--radius-sm);
		color: var(--color-text);
		font-size: 0.8125rem;
		text-align: left;
		cursor: pointer;
		transition:
			background-color 0.15s ease,
			color 0.15s ease;
	}

	button:hover {
		background: var(--color-bg-hover);
	}

	button.active {
		background: var(--color-bg-active);
	}

	.icon {
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.error {
		color: var(--color-error);
	}

	.done {
		color: var(--color-success);
	}

	.skip {
		color: var(--color-skip);
	}

	.running {
		color: var(--color-running);
		animation: pulse 1.5s ease-in-out infinite;
	}

	.name {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.duration {
		flex-shrink: 0;
		font-size: 0.75rem;
		color: var(--color-text-muted);
		font-variant-numeric: tabular-nums;
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.5;
		}
	}
</style>
