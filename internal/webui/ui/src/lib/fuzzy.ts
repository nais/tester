/**
 * Simple fuzzy search implementation that matches a pattern against text
 * Returns a score (higher is better) if there's a match, or -1 if no match
 */
export function fuzzyMatch(pattern: string, text: string): number {
	// Convert to lowercase for case-insensitive matching
	const patternLower = pattern.toLowerCase();
	const textLower = text.toLowerCase();

	// If pattern is empty, match everything with score 0
	if (patternLower.length === 0) {
		return 0;
	}

	// If text doesn't contain pattern at all, no match
	if (!textLower.includes(patternLower)) {
		let patternIdx = 0;
		let textIdx = 0;

		// Try character-by-character fuzzy matching
		while (textIdx < textLower.length && patternIdx < patternLower.length) {
			if (textLower[textIdx] === patternLower[patternIdx]) {
				patternIdx++;
			}
			textIdx++;
		}

		// If we didn't match all pattern characters, no match
		if (patternIdx !== patternLower.length) {
			return -1;
		}
	}

	// Calculate score based on match quality
	let score = 0;
	const exactMatch = textLower === patternLower;
	const startsWithMatch = textLower.startsWith(patternLower);
	const wordMatch = textLower.split(/[_\-/.]/).some((part) => part.startsWith(patternLower));

	if (exactMatch) {
		score = 1000;
	} else if (startsWithMatch) {
		score = 500;
	} else if (wordMatch) {
		score = 250;
	} else {
		score = 100;
	}

	// Boost score for shorter strings (more specific matches)
	score += Math.max(0, 100 - text.length);

	return score;
}

/**
 * Fuzzy search through a list of items
 */
export function fuzzySearch<T>(items: T[], pattern: string, getText: (item: T) => string): T[] {
	if (!pattern) {
		return items;
	}

	const matches = items
		.map((item) => ({
			item,
			score: fuzzyMatch(pattern, getText(item)),
		}))
		.filter((result) => result.score >= 0)
		.sort((a, b) => b.score - a.score);

	return matches.map((m) => m.item);
}
