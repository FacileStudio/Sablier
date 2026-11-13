import { describe, expect, it, vi } from 'vitest';

vi.mock('$lib/i18n/index.svelte', () => ({ t: (key: string) => key }));

const { APK_SHA256_URL, DOWNLOADS, INSTALL_METHODS } = await import('./downloads');

const LATEST = 'https://github.com/FacileStudio/Sablier/releases/latest/download';

describe('release assets', () => {
	it('point at the files the desktop workflow publishes', () => {
		expect(DOWNLOADS.map((entry) => entry.href)).toEqual([
			`${LATEST}/Sablier-macos-universal.dmg`,
			`${LATEST}/Sablier-windows-x64-setup.exe`,
			`${LATEST}/Sablier-linux-x64.deb`,
			`${LATEST}/Sablier-linux-x64.rpm`,
			`${LATEST}/Sablier-android-universal.apk`
		]);
		expect(APK_SHA256_URL).toBe(`${LATEST}/Sablier-android-universal.apk.sha256`);
	});

	it('install the CLI from sablier-cli and the FacileStudio tap', () => {
		expect(INSTALL_METHODS.map((method) => method.command)).toEqual([
			'curl -fsSL https://raw.githubusercontent.com/FacileStudio/sablier-cli/main/install.sh | bash',
			'facile install sablier',
			'brew install FacileStudio/tap/sablier'
		]);
	});
});
