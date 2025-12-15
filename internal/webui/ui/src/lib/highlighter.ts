import { createHighlighter, type Highlighter } from "shiki";

let highlighter: Highlighter | null = null;
let highlighterPromise: Promise<Highlighter> | null = null;

const SUPPORTED_LANGUAGES = ["go", "lua", "diff", "graphql", "sql", "json"] as const;

async function getHighlighter(): Promise<Highlighter> {
	if (highlighter) {
		return highlighter;
	}

	if (!highlighterPromise) {
		highlighterPromise = createHighlighter({
			themes: ["github-dark"],
			langs: [...SUPPORTED_LANGUAGES],
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
