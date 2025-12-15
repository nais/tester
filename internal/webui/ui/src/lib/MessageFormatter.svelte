<script lang="ts">
	import CodeView from "./CodeView.svelte";
	import type { TestError } from "./watcher.svelte";

	let { error }: { error: TestError } = $props();

	type Tab = "diff" | "expected" | "actual";
	let activeTab: Tab = $state("diff");

	const hasStructuredData = $derived(error.expected !== undefined || error.actual !== undefined);

	// Convert a value to Lua table syntax
	function toLua(value: unknown, indent: number = 0): string {
		const pad = "\t".repeat(indent);
		const padInner = "\t".repeat(indent + 1);

		if (value === null || value === undefined) {
			return "Null";
		}

		if (typeof value === "string") {
			// Check for special placeholder strings
			if (value === "[[[ save ]]]" || value === "[[[ save_allow_null ]]]") {
				return "Save(...)";
			}
			if (value === "[[[ ignore ]]]") {
				return "Ignore()";
			}
			if (value === "[[[ not_null ]]]") {
				return "NotNull()";
			}
			if (value === "[[[ empty_list_or_map ]]]") {
				return "{}";
			}
			if (value.startsWith("[[[ contains")) {
				return "Contains(...)";
			}
			// Escape quotes and backslashes
			const escaped = value
				.replace(/\\/g, "\\\\")
				.replace(/"/g, '\\"')
				.replace(/\n/g, "\\n")
				.replace(/\t/g, "\\t");
			return `"${escaped}"`;
		}

		if (typeof value === "number") {
			return String(value);
		}

		if (typeof value === "boolean") {
			return value ? "true" : "false";
		}

		if (Array.isArray(value)) {
			if (value.length === 0) {
				return "{}";
			}
			const items = value.map((item) => `${padInner}${toLua(item, indent + 1)}`).join(",\n");
			return `{\n${items},\n${pad}}`;
		}

		if (typeof value === "object") {
			const entries = Object.entries(value);
			if (entries.length === 0) {
				return "{}";
			}
			const items = entries
				.map(([key, val]) => {
					// Use bracket notation for keys that aren't valid identifiers
					const keyStr = /^[a-zA-Z_][a-zA-Z0-9_]*$/.test(key) ? key : `["${key}"]`;
					return `${padInner}${keyStr} = ${toLua(val, indent + 1)}`;
				})
				.join(",\n");
			return `{\n${items},\n${pad}}`;
		}

		return String(value);
	}

	const expectedLua = $derived(toLua(error.expected));
	const actualLua = $derived(toLua(error.actual));
</script>

<div class="message-block">
	{#if hasStructuredData}
		<div class="tabs">
			<button class="tab" class:active={activeTab === "diff"} onclick={() => (activeTab = "diff")}>
				Diff
			</button>
			<button
				class="tab"
				class:active={activeTab === "expected"}
				onclick={() => (activeTab = "expected")}
			>
				Expected
			</button>
			<button
				class="tab"
				class:active={activeTab === "actual"}
				onclick={() => (activeTab = "actual")}
			>
				Actual
			</button>
		</div>
	{/if}

	<div class="tab-content">
		{#if activeTab === "diff"}
			<CodeView code={error.message} lang="diff" />
		{:else if activeTab === "expected"}
			<CodeView code={expectedLua} lang="lua" />
		{:else if activeTab === "actual"}
			<CodeView code={actualLua} lang="lua" />
		{/if}
	</div>
</div>

<style>
	.message-block {
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		overflow: hidden;
	}

	.tabs {
		display: flex;
		gap: 0;
		border-bottom: 1px solid var(--color-border);
		background: var(--color-bg);
	}

	.tab {
		padding: 0.5rem 1rem;
		font-size: 0.8125rem;
		background: transparent;
		border: none;
		border-bottom: 2px solid transparent;
		color: var(--color-text-muted);
		cursor: pointer;
		transition:
			color 0.15s ease,
			border-color 0.15s ease;
	}

	.tab:hover {
		color: var(--color-text);
	}

	.tab.active {
		color: var(--color-text);
		border-bottom-color: var(--color-text);
	}

	.tab-content {
		overflow: hidden;
	}
</style>
