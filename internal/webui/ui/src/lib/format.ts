export function formatNanoseconds(ns: number) {
	if (ns < 1e9) {
		return `${(ns / 1e6).toFixed(2)}ms`;
	} else {
		return `${(ns / 1e9).toFixed(2)}s`;
	}
}
