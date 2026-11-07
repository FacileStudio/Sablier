export type StepId = 'everywhere' | 'forget' | 'agents' | 'signin';

export type Step = {
	id: StepId;
};

export type Landing = {
	native: boolean;
	signedIn: boolean;
};

export const ONBOARDED_KEY = 'sablier.onboarded';

export const STEPS: readonly Step[] = [
	{ id: 'everywhere' },
	{ id: 'forget' },
	{ id: 'agents' },
	{ id: 'signin' }
];

/** Keeps a step index inside the step list. */
export function clampStep(index: number): number {
	return Math.min(Math.max(Math.trunc(index), 0), STEPS.length - 1);
}

/** Reports whether the index is the sign-in step. */
export function isLastStep(index: number): boolean {
	return index >= STEPS.length - 1;
}

/** Returns the step to open: the sign-in step for people who already went through onboarding. */
export function startStep(onboarded: boolean): number {
	return onboarded ? STEPS.length - 1 : 0;
}

/** Returns how much sand has fallen at a step, from 0 to 1. */
export function drained(index: number): number {
	return (clampStep(index) + 1) / STEPS.length;
}

/** Decides where `/` sends a visitor; null keeps the landing page. */
export function landingTarget({ native, signedIn }: Landing): '/dashboard' | '/welcome' | null {
	if (signedIn) return '/dashboard';
	return native ? '/welcome' : null;
}

/** Reads the completion flag; storage failures count as not onboarded. */
export function readOnboarded(store?: Pick<Storage, 'getItem'>): boolean {
	try {
		return (store ?? localStorage).getItem(ONBOARDED_KEY) === '1';
	} catch {
		return false;
	}
}

/** Stores the completion flag, ignoring storage failures. */
export function markOnboarded(store?: Pick<Storage, 'setItem'>): void {
	try {
		(store ?? localStorage).setItem(ONBOARDED_KEY, '1');
	} catch {
		return;
	}
}
