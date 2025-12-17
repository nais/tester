import { createHighlighterCore, type HighlighterCore } from "shiki/core";
import { createJavaScriptRegexEngine } from "shiki/engine/javascript";

let highlighter: HighlighterCore | null = null;
let highlighterPromise: Promise<HighlighterCore> | null = null;

const SUPPORTED_LANGUAGES = ["go", "lua", "diff", "graphql", "sql", "json"] as const;

async function getHighlighter(): Promise<HighlighterCore> {
	if (highlighter) {
		return highlighter;
	}

	if (!highlighterPromise) {
		highlighterPromise = createHighlighterCore({
			themes: [import("shiki/themes/github-dark.mjs")],
			langs: [
				import("shiki/langs/go.mjs"),
				import("shiki/langs/lua.mjs"),
				import("shiki/langs/diff.mjs"),
				import("shiki/langs/graphql.mjs"),
				import("shiki/langs/sql.mjs"),
				import("shiki/langs/json.mjs"),
			],
			engine: createJavaScriptRegexEngine(),
		}).then((h) => {
			highlighter = h;
			return h;
		});
	}

	return highlighterPromise;
}

export type SupportedLanguages = (typeof SUPPORTED_LANGUAGES)[number];

export async function highlightCode(code: string, lang: SupportedLanguages): Promise<string> {
	const h = await getHighlighter();
	return h.codeToHtml(code, {
		lang,
		theme: "github-dark",
	});
}
