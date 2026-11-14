import { registerIcons } from '@facile/muse';
import data from './landing-icons.json';

/** The Solar glyphs the landing draws that muse does not bundle, so none of them waits on the network. */
export const LANDING_ICONS = {
	sablier: 'solar:hourglass-bold-duotone',
	opus: 'solar:slider-vertical-minimalistic-bold-duotone',
	bulle: 'solar:smile-circle-bold-duotone',
	play: 'solar:play-linear',
	pause: 'solar:pause-linear',
	stop: 'solar:stop-linear',
	phone: 'solar:smartphone-2-linear',
	laptop: 'solar:laptop-minimalistic-linear',
	synced: 'solar:check-circle-bold',
	wallet: 'solar:wallet-money-linear'
} as const;

registerIcons(data);
