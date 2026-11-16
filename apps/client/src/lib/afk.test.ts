import { describe, expect, it } from 'vitest';
import { STILL_WORKING_AFTER_MS, stillWorkingDue, stopAtOffer } from './afk';

const start = '2026-11-16T08:00:00Z';
const at = (ms: number) => Date.parse(start) + ms;

describe('stillWorkingDue', () => {
	it('asks once the timer has been idle for four hours', () => {
		const entry = { started_at: start, paused_at: null, last_active_at: start };
		expect(stillWorkingDue(entry, at(STILL_WORKING_AFTER_MS - 1))).toBe(false);
		expect(stillWorkingDue(entry, at(STILL_WORKING_AFTER_MS))).toBe(true);
	});

	it('measures from the last activity, not the start', () => {
		const entry = { started_at: start, paused_at: null, last_active_at: '2026-11-16T11:00:00Z' };
		expect(stillWorkingDue(entry, at(STILL_WORKING_AFTER_MS))).toBe(false);
	});

	it('never asks about a paused timer or while snoozed', () => {
		const paused = { started_at: start, paused_at: start, last_active_at: start };
		expect(stillWorkingDue(paused, at(STILL_WORKING_AFTER_MS * 3))).toBe(false);
		const running = { ...paused, paused_at: null };
		expect(stillWorkingDue(running, at(STILL_WORKING_AFTER_MS), at(STILL_WORKING_AFTER_MS + 1))).toBe(false);
	});
});

describe('stopAtOffer', () => {
	it('offers the last activity only when it follows the start', () => {
		expect(stopAtOffer({ started_at: start, paused_at: null, last_active_at: start })).toBeNull();
		expect(stopAtOffer({ started_at: start, paused_at: null, last_active_at: '2026-11-16T09:30:00Z' })).toBe('2026-11-16T09:30:00Z');
	});
});
