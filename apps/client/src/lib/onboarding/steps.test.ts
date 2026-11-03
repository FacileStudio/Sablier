import { describe, expect, test } from 'vitest';
import { STEPS, clampStep, drained, landingTarget, markOnboarded, readOnboarded, startStep } from './steps';

function memoryStore(): Pick<Storage, 'getItem' | 'setItem'> {
	const values = new Map<string, string>();
	return {
		getItem: (key) => values.get(key) ?? null,
		setItem: (key, value) => void values.set(key, value)
	};
}

const broken: Pick<Storage, 'getItem' | 'setItem'> = {
	getItem: () => {
		throw new Error('denied');
	},
	setItem: () => {
		throw new Error('denied');
	}
};

describe('onboarding steps', () => {
	test('run in order and end on sign in', () => {
		expect(STEPS.map((s) => s.id)).toEqual(['everywhere', 'forget', 'agents', 'signin']);
	});

	test('resume a returning visitor at sign in and clamp out-of-range steps', () => {
		expect(startStep(false)).toBe(0);
		expect(startStep(true)).toBe(STEPS.length - 1);
		expect(clampStep(-3)).toBe(0);
		expect(clampStep(99)).toBe(STEPS.length - 1);
	});

	test('drain the hourglass a little each step until it is empty at sign in', () => {
		expect(drained(0)).toBeGreaterThan(0);
		expect(drained(0)).toBeLessThan(drained(1));
		expect(drained(STEPS.length - 1)).toBe(1);
	});
});

describe('landingTarget', () => {
	test('sends signed-in users to the dashboard and signed-out native users to onboarding', () => {
		expect(landingTarget({ native: true, signedIn: true })).toBe('/dashboard');
		expect(landingTarget({ native: true, signedIn: false })).toBe('/welcome');
		expect(landingTarget({ native: false, signedIn: true })).toBe('/dashboard');
		expect(landingTarget({ native: false, signedIn: false })).toBeNull();
	});
});

describe('onboarded flag', () => {
	test('is stored and survives a storage failure as not onboarded', () => {
		const store = memoryStore();
		expect(readOnboarded(store)).toBe(false);
		markOnboarded(store);
		expect(readOnboarded(store)).toBe(true);
		markOnboarded(broken);
		expect(readOnboarded(broken)).toBe(false);
	});
});
