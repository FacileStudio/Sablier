import type { Catalogs, Message } from '@facile/muse';
import type { Row } from '../row';
import { shell } from './shell';
import { dashboard } from './dashboard';
import { projects } from './projects';
import { settings } from './settings';
import { timer } from './timer';
import { ideas } from './ideas';
import { pending } from './pending';
import { landing } from './landing';
import { landingSections } from './landing-sections';

const rows = {
	...shell,
	...dashboard,
	...projects,
	...settings,
	...timer,
	...ideas,
	...pending,
	...landing,
	...landingSections
};

export type AppKey = keyof typeof rows;
type AppCatalog = Record<AppKey, Message>;

function column(index: number): AppCatalog {
	const entries = Object.entries(rows as Record<string, Row>).map(([key, row]) => [key, row[index]]);
	return Object.fromEntries(entries) as AppCatalog;
}

/** The app's catalogs: English is the typed base, and every row carries all four locales. */
export const catalogs: Catalogs<AppCatalog> = {
	en: column(0),
	fr: column(1),
	es: column(2),
	de: column(3)
};
