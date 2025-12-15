<script lang="ts">
	let { message }: { message: string } = $props();

	const lines = $derived(
		message.split("\n").map((line, i) => ({
			line: i,
			m: line,
			add: line.startsWith("+"),
			del: line.startsWith("-"),
		})),
	);

	const maxLines = $derived(lines.length);

	const padZero = (num: number) => num.toString().padStart(maxLines.toString().length, "0");
</script>

<div class="message-block">
	<pre>{#each lines as { m, add, del, line } (line)}<span
				class:add
				class:del
				data-line={padZero(line)}
				>{m}
</span>{/each}</pre>
</div>

<style>
	.message-block {
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		overflow: hidden;
	}

	pre {
		margin: 0;
		padding: 0.75rem 0;
		white-space: pre-wrap;
		font-family: ui-monospace, "SF Mono", Menlo, Monaco, "Cascadia Code", monospace;
		font-size: 0.8125rem;
		line-height: 1.6;
		overflow-x: auto;
	}

	span {
		display: block;
		padding: 0 1rem 0 3.5rem;
		position: relative;
	}

	span:hover {
		background: var(--color-bg-hover);
	}

	span::before {
		content: attr(data-line);
		position: absolute;
		left: 0;
		width: 2.5rem;
		padding-right: 0.5rem;
		text-align: right;
		color: var(--color-text-muted);
		user-select: none;
	}

	.add {
		background: color-mix(in srgb, var(--color-success) 10%, transparent);
		color: var(--color-success);
	}

	.add:hover {
		background: color-mix(in srgb, var(--color-success) 18%, transparent);
	}

	.del {
		background: color-mix(in srgb, var(--color-error) 10%, transparent);
		color: var(--color-error);
	}

	.del:hover {
		background: color-mix(in srgb, var(--color-error) 18%, transparent);
	}
</style>
