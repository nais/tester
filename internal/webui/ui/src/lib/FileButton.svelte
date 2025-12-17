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
		showRerun = false,
	}: {
		file: { name: string; status: Status; duration: number };
		active: boolean;
		onselect: (name: string) => void;
		showRerun?: boolean;
	} = $props();

	let rerunning = $state(false);

	async function rerun(e: MouseEvent) {
		e.stopPropagation();
		if (rerunning) return;

		rerunning = true;
		try {
			await fetch(`/rerun?file=${encodeURIComponent(file.name)}`, {
				method: "POST",
			});
		} finally {
			// Keep the rerunning state until the file status changes to RUNNING
			// or after a timeout
			setTimeout(() => {
				rerunning = false;
			}, 2000);
		}
	}
</script>

<div class="file-row" class:active>
	<button class="file-btn" onclick={() => onselect(file.name)}>
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
	{#if showRerun}
		<button
			class="rerun-btn"
			onclick={rerun}
			disabled={rerunning || file.status === Status.RUNNING}
			title="Rerun test file"
		>
			{#if rerunning || file.status === Status.RUNNING}
				⏳
			{:else}
				▶
			{/if}
		</button>
	{/if}
</div>

<style>
	.file-row {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		padding-right: 0.5rem;
		border-radius: var(--radius-sm);
		transition: background-color 0.15s ease;
	}

	.file-row:hover {
		background: var(--color-bg-hover);
	}

	.file-row.active {
		background: var(--color-bg-active);
	}

	.file-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex: 1;
		min-width: 0;
		padding: 0.5rem 0.5rem 0.5rem 0.75rem;
		background: transparent;
		border: none;
		color: var(--color-text);
		font-size: 0.8125rem;
		text-align: left;
		cursor: pointer;
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

	.rerun-btn {
		flex-shrink: 0;
		width: 1.5rem;
		height: 1.5rem;
		padding: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		font-size: 0.625rem;
		color: var(--color-text);
		cursor: pointer;
		opacity: 0;
		transition:
			opacity 0.15s ease,
			background 0.15s ease;
	}

	.file-row:hover .rerun-btn {
		opacity: 0.7;
	}

	.rerun-btn:hover {
		opacity: 1 !important;
		background: var(--color-bg-hover);
	}

	.rerun-btn:disabled {
		cursor: not-allowed;
		opacity: 0.5 !important;
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
