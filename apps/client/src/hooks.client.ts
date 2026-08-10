import { handleErrorWith } from '@facile/journal/sveltekit';

import { journal } from '$lib/journal';

/**
 * install() catches what reaches the window — an uncaught throw, a rejected
 * promise. handleError catches what SvelteKit swallows first: a load function
 * that threw, a component that failed to render. Those never reach the window,
 * so both are needed.
 */
journal.install();

export const handleError = handleErrorWith(journal);
