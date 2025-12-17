<script lang="ts">
	interface Props {
		onclick: () => void;
	}

	let { onclick }: Props = $props();

	// Detect OS to show correct keybind
	const isMac =
		typeof navigator !== "undefined" &&
		/(Mac|iPhone|iPod|iPad)/i.test(navigator.platform || navigator.userAgent);

	const keybind = isMac ? "⌘K" : "Ctrl+K";
</script>

<button class="command-button" {onclick} type="button" title="Open command panel ({keybind})">
	<span class="icon">⌘</span>
	<span class="label">Command Panel</span>
	<kbd>{keybind}</kbd>
</button>

<style>
	.command-button {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.75rem;
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		color: var(--color-text);
		font-size: 0.875rem;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.command-button:hover {
		background: var(--color-bg-hover);
		border-color: var(--color-running);
	}

	.icon {
		font-size: 1rem;
		color: var(--color-text-muted);
	}

	.label {
		flex: 1;
	}

	kbd {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 0.125rem 0.375rem;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm, 4px);
		font-family: ui-monospace, monospace;
		font-size: 0.625rem;
		font-weight: 600;
		color: var(--color-text-muted);
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
	}
</style>
