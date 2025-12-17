<script lang="ts">
	import { formatNanoseconds } from "./format";
	import type { TestInfo } from "./watcher.svelte";

	let { info }: { info: TestInfo } = $props();

	let expanded = $state(false);

	const hasArgs = $derived(info.args && info.args.length > 0);
	const hasLongContent = $derived(
		hasArgs ||
			info.content.length > 60 ||
			(info.args && info.args.some((a) => a.value.length > 50)),
	);

	// Format args for the collapsed view
	const collapsedContent = $derived.by(() => {
		if (info.args && info.args.length > 0) {
			return info.args.map((arg) => (arg.name ? `${arg.name}=${arg.value}` : arg.value)).join(", ");
		}
		return info.content;
	});
</script>

<div class="info-row-wrapper" class:expanded>
	<button class="info-row" onclick={() => (expanded = !expanded)}>
		<span class="info-icon">⚙</span>
		<span class="info-title">{info.title}</span>
		<span class="info-content">{collapsedContent}</span>
		<span class="info-timestamp">{formatNanoseconds(info.timestamp)}</span>
		{#if hasLongContent}
			<span class="expand-indicator">{expanded ? "▼" : "▶"}</span>
		{/if}
	</button>

	{#if expanded && hasLongContent}
		<div class="info-expanded">
			{#if hasArgs}
				<table class="args-table">
					<tbody>
						{#each info.args as arg, i (i)}
							<tr>
								<td class="arg-key">{arg.name || `[${i}]`}</td>
								<td class="arg-value">{arg.value}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{:else if info.content}
				<pre class="raw-content">{info.content}</pre>
			{/if}
		</div>
	{/if}
</div>

<style>
	.info-row-wrapper {
		border-bottom: 1px solid var(--color-border);
	}

	.info-row-wrapper.expanded {
		background: var(--color-bg-elevated);
	}

	.info-row {
		display: flex;
		flex-direction: row;
		align-items: center;
		gap: 0.5rem;
		padding: 0.375rem 0.75rem;
		font-size: 0.75rem;
		color: var(--color-text-muted);
		background: transparent;
		border: none;
		width: 100%;
		text-align: left;
		cursor: pointer;
	}

	.info-row:hover {
		background: var(--color-bg-hover);
	}

	.info-row .info-icon {
		width: 1rem;
		flex-shrink: 0;
		opacity: 0.6;
	}

	.info-row .info-title {
		font-weight: 500;
		color: var(--color-text);
		flex-shrink: 0;
	}

	.info-row .info-content {
		flex: 1;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		opacity: 0.7;
		font-family: monospace;
	}

	.info-row .info-timestamp {
		font-size: 0.6875rem;
		color: var(--color-text-muted);
		flex-shrink: 0;
	}

	.expand-indicator {
		font-size: 0.625rem;
		color: var(--color-text-muted);
		flex-shrink: 0;
		width: 1rem;
		text-align: center;
	}

	.info-expanded {
		padding: 0.5rem 0.75rem 0.75rem 2.25rem;
		border-top: 1px solid var(--color-border);
	}

	.args-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.75rem;
	}

	.args-table tr:not(:last-child) td {
		border-bottom: 1px solid var(--color-border);
	}

	.arg-key {
		padding: 0.375rem 0.5rem 0.375rem 0;
		color: var(--color-text-muted);
		font-weight: 500;
		white-space: nowrap;
		vertical-align: top;
		width: 1%;
	}

	.arg-value {
		padding: 0.375rem 0;
		font-family: monospace;
		color: var(--color-text);
		word-break: break-all;
		white-space: pre-wrap;
	}

	.raw-content {
		margin: 0;
		padding: 0.5rem;
		background: var(--color-bg);
		border-radius: var(--radius-sm);
		font-family: monospace;
		font-size: 0.75rem;
		white-space: pre-wrap;
		word-break: break-all;
		color: var(--color-text);
	}
</style>
