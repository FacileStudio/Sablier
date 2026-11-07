import { describe, expect, test } from 'vitest';
import { PROJECT_ICONS, toIconify } from './icons';

describe('project icons', () => {
	test('offers every linear Solar icon once', () => {
		expect(PROJECT_ICONS.length).toBe(1451);
		expect(new Set(PROJECT_ICONS).size).toBe(PROJECT_ICONS.length);
	});

	test('maps names to their exact Solar id', () => {
		expect(toIconify('AddCircle')).toBe('solar:add-circle-linear');
		expect(toIconify('Grid2x2')).toBe('solar:grid-2x2-linear');
		expect(toIconify('Magnifier')).toBe('solar:magnifier-linear');
		expect(toIconify('CardReceive')).toBe('solar:card-receive-linear');
	});

	test('keeps names Solar renamed and the empty default working', () => {
		expect(toIconify('Accumulator')).toBe('solar:accumulator-linear');
		expect(toIconify(null)).toBe('solar:widget-2-linear');
	});
});
