<script lang="ts">
	import { highlightCode, type SupportedLanguages } from "./highlighter";

	let { code, lang }: { code: string; lang: SupportedLanguages } = $props();

	let highlightedHtml = $state("");
	let prevCode = "";
	let copied = $state(false);

	$effect(() => {
		if (code !== prevCode) {
			prevCode = code;
			highlightedHtml = "";
		}

		if (code && !highlightedHtml) {
			highlightCode(code, lang).then((html) => {
				highlightedHtml = html;
			});
		}
	});

	async function copyToClipboard() {
		await navigator.clipboard.writeText(code);
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}
</script>

<div class="code-view">
	<button class="copy-btn" onclick={copyToClipboard} title="Copy code">
		{copied ? "✓" : "📋"}
	</button>
	{#if highlightedHtml}
		<!-- eslint-disable-next-line svelte/no-at-html-tags -->
		{@html highlightedHtml}
	{:else}
		<pre>{code}</pre>
	{/if}
</div>

<style>
	.code-view {
		position: relative;
	}

	.copy-btn {
		position: absolute;
		top: 0.5rem;
		right: 0.5rem;
		padding: 0.375rem 0.5rem;
		font-size: 0.875rem;
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		color: var(--color-text);
		cursor: pointer;
		opacity: 0.7;
		transition:
			opacity 0.15s ease,
			background 0.15s ease;
		z-index: 1;
	}

	.copy-btn:hover {
		opacity: 1;
		background: var(--color-bg-hover);
	}

	.copy-btn:active {
		background: var(--color-bg-active);
	}

	pre {
		margin: 0;
		padding: 0.75rem;
		white-space: pre-wrap;
		font-family: ui-monospace, "SF Mono", Menlo, Monaco, "Cascadia Code", monospace;
		font-size: 0.8125rem;
		line-height: 1.6;
		overflow-x: auto;
	}
</style>
