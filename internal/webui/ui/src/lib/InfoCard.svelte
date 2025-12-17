<script lang="ts">
	import CodeView from "./CodeView.svelte";
	import { formatNanoseconds } from "./format";
	import type { TestInfo } from "./watcher.svelte";

	let { info }: { info: TestInfo } = $props();

	const iconMap: Record<string, string> = {
		helper: "⚙️",
		request: "📤",
		response: "📥",
		query: "🔍",
		result: "📋",
	};

	const colorMap: Record<string, string> = {
		helper: "var(--color-info-helper)",
		request: "var(--color-info-request)",
		response: "var(--color-info-response)",
		query: "var(--color-info-query)",
		result: "var(--color-info-result)",
	};

	let expanded = $state(false);
	const hasArgs = $derived(info.args && info.args.length > 0);
	const isLongContent = $derived(
		hasArgs || info.content.length > 200 || info.content.split("\n").length > 5,
	);
</script>

<div class="info-card" style:--accent-color={colorMap[info.type] ?? "var(--color-text-muted)"}>
	<div class="header">
		<span class="icon">{iconMap[info.type] ?? "ℹ️"}</span>
		<span class="type-badge">{info.type}</span>
		<span class="title">{info.title}</span>
		<span class="timestamp">{formatNanoseconds(info.timestamp)}</span>
	</div>
	{#if isLongContent && !expanded}
		<button class="collapsed-card" onclick={() => (expanded = true)}>
			<div class="content-wrapper collapsed">
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
				{/if}
				{#if info.content}
					{#if info.language}
						<CodeView code={info.content} lang={info.language} />
					{:else}
						<pre class="content">{info.language}{info.content}</pre>
					{/if}
				{/if}
			</div>
			<div class="expand-hint">
				<span>Click to expand</span>
				<span class="expand-icon">▼</span>
			</div>
		</button>
	{:else}
		<div class="content-wrapper">
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
			{/if}
			{#if info.content}
				{#if info.language}
					<CodeView code={info.content} lang={info.language} />
				{:else}
					<pre class="content">{info.language}{info.content}</pre>
				{/if}
			{/if}
		</div>
		{#if isLongContent}
			<button class="collapse-button" onclick={() => (expanded = false)}>
				<span>Collapse</span>
				<span class="collapse-icon">▲</span>
			</button>
		{/if}
	{/if}
</div>

<style>
	.info-card {
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		background: var(--color-bg-elevated);
		overflow: hidden;
		border-left: 3px solid var(--accent-color);
	}

	.header {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.75rem;
		background: transparent;
		border-bottom: 1px solid var(--color-border);
		color: var(--color-text);
		font-size: 0.875rem;
	}

	.icon {
		font-size: 1rem;
	}

	.type-badge {
		font-size: 0.7rem;
		padding: 0.125rem 0.375rem;
		background: var(--accent-color);
		color: var(--color-bg);
		border-radius: 3px;
		text-transform: uppercase;
		font-weight: 600;
	}

	.title {
		flex: 1;
		font-weight: 500;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.timestamp {
		font-size: 0.6875rem;
		color: var(--color-text-muted);
		flex-shrink: 0;
	}

	.collapsed-card {
		width: 100%;
		background: transparent;
		border: none;
		cursor: pointer;
		padding: 0;
		text-align: left;
		color: var(--color-text);
		transition: background-color 0.15s ease;
	}

	.collapsed-card:hover {
		background: var(--color-bg-hover);
	}

	.collapsed-card:hover .expand-hint {
		background: var(--color-bg-active);
	}

	.expand-hint {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		padding: 0.5rem;
		background: var(--color-bg);
		border-top: 1px solid var(--color-border);
		font-size: 0.75rem;
		color: var(--color-text-muted);
		transition: background-color 0.15s ease;
	}

	.expand-icon {
		font-size: 0.625rem;
	}

	.collapse-button {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		width: 100%;
		padding: 0.5rem;
		background: var(--color-bg);
		border: none;
		border-top: 1px solid var(--color-border);
		cursor: pointer;
		font-size: 0.75rem;
		color: var(--color-text-muted);
		transition: background-color 0.15s ease;
	}

	.collapse-button:hover {
		background: var(--color-bg-active);
		color: var(--color-text);
	}

	.collapse-icon {
		font-size: 0.625rem;
	}

	.content-wrapper {
		background: var(--color-bg);
	}

	.collapsed-card .content-wrapper.collapsed {
		max-height: 100px;
		overflow: hidden;
		position: relative;
		pointer-events: none;
	}

	.collapsed-card .content-wrapper.collapsed::after {
		content: "";
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		height: 40px;
		background: linear-gradient(transparent, var(--color-bg));
	}

	.args-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.8rem;
	}

	.args-table tr:not(:last-child) td {
		border-bottom: 1px solid var(--color-border);
	}

	.arg-key {
		padding: 0.5rem 0.75rem;
		color: var(--color-text-muted);
		font-weight: 500;
		font-family: monospace;
		white-space: nowrap;
		vertical-align: top;
		width: 1%;
	}

	.arg-value {
		padding: 0.5rem 0.75rem;
		font-family: monospace;
		color: var(--color-text);
		word-break: break-all;
		white-space: pre-wrap;
	}

	.content {
		margin: 0;
		padding: 0.75rem;
		font-family: monospace;
		font-size: 0.8rem;
		white-space: pre-wrap;
		word-break: break-word;
		color: var(--color-text);
		overflow-x: auto;
	}
</style>
