import {
	STEPS,
	clampStep,
	drained,
	landingTarget,
	markOnboarded,
	readOnboarded,
	startStep
} from './steps';

/** Throws with the label when the two values differ. */
function equal(actual: unknown, expected: unknown, label: string): void {
	if (JSON.stringify(actual) !== JSON.stringify(expected)) {
		throw new Error(`${label}: expected ${JSON.stringify(expected)}, got ${JSON.stringify(actual)}`);
	}
}

/** Builds an in-memory stand-in for localStorage. */
function memoryStore(): Pick<Storage, 'getItem' | 'setItem'> {
	const values = new Map<string, string>();
	return {
		getItem: (key) => values.get(key) ?? null,
		setItem: (key, value) => void values.set(key, value)
	};
}

const broken = {
	getItem: (): string | null => {
		throw new Error('denied');
	},
	setItem: (): void => {
		throw new Error('denied');
	}
};

equal(
	STEPS.map((s) => s.id),
	['everywhere', 'forget', 'agents', 'signin'],
	'step order'
);
equal(startStep(false), 0, 'first visit starts at the first step');
equal(startStep(true), STEPS.length - 1, 'returning visitor resumes at sign in');
equal(clampStep(-3), 0, 'clamp low');
equal(clampStep(99), STEPS.length - 1, 'clamp high');
equal(drained(0) > 0 && drained(0) < drained(1), true, 'sand drains a little each step');
equal(drained(STEPS.length - 1), 1, 'sand is gone at sign in');

equal(landingTarget({ native: true, signedIn: true }), '/dashboard', 'native signed in');
equal(landingTarget({ native: true, signedIn: false }), '/welcome', 'native signed out');
equal(landingTarget({ native: false, signedIn: true }), '/dashboard', 'web signed in');
equal(landingTarget({ native: false, signedIn: false }), null, 'web signed out keeps the landing');

const store = memoryStore();
equal(readOnboarded(store), false, 'not onboarded by default');
markOnboarded(store);
equal(readOnboarded(store), true, 'onboarded after marking');
markOnboarded(broken);
equal(readOnboarded(broken), false, 'storage failure reads as not onboarded');

console.log('onboarding steps: ok');
