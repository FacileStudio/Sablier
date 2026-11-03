import { describe, expect, it } from 'vitest';
import { APK_SHA256_URL, APK_URL, DESKTOP_BUILDS, detectOS, isAndroid } from './downloads';

const LATEST = 'https://github.com/FacileStudio/Sablier/releases/latest/download';

describe('release assets', () => {
	it('point at the files the desktop workflow publishes', () => {
		expect(DESKTOP_BUILDS.map((build) => build.url)).toEqual([
			`${LATEST}/Sablier-macos-universal.dmg`,
			`${LATEST}/Sablier-linux-x64.deb`,
			`${LATEST}/Sablier-linux-x64.rpm`
		]);
		expect(APK_URL).toBe(`${LATEST}/Sablier-android-universal.apk`);
		expect(APK_SHA256_URL).toBe(`${LATEST}/Sablier-android-universal.apk.sha256`);
	});
});

describe('detectOS', () => {
	it('recognises desktops and leaves phones alone', () => {
		expect(detectOS('Mozilla/5.0 (Macintosh; Intel Mac OS X 14_5)')).toBe('macos');
		expect(detectOS('Mozilla/5.0 (X11; Linux x86_64)')).toBe('linux');
		expect(detectOS('Mozilla/5.0 (Linux; Android 14; Pixel 8)')).toBeNull();
		expect(detectOS('Mozilla/5.0 (Windows NT 10.0; Win64; x64)')).toBeNull();
	});

	it('spots android', () => {
		expect(isAndroid('Mozilla/5.0 (Linux; Android 14; Pixel 8)')).toBe(true);
		expect(isAndroid('Mozilla/5.0 (X11; Linux x86_64)')).toBe(false);
	});
});
