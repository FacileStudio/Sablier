const RELEASES = 'https://github.com/FacileStudio/Sablier/releases';
const LATEST = `${RELEASES}/latest/download`;

export const APK_ASSET = 'Sablier-android-universal.apk';
export const APK_URL = `${LATEST}/${APK_ASSET}`;
export const APK_SHA256_URL = `${APK_URL}.sha256`;
export const RELEASES_URL = RELEASES;
export const DESKTOP_CHECKSUMS_URL = `${LATEST}/desktop-checksums.txt`;

export type DesktopOS = 'macos' | 'linux';

export type DesktopBuild = { id: 'dmg' | 'deb' | 'rpm'; os: DesktopOS; url: string };

export const DESKTOP_BUILDS: readonly DesktopBuild[] = [
	{
		id: 'dmg',
		os: 'macos',
		url: `${LATEST}/Sablier-macos-universal.dmg`
	},
	{
		id: 'deb',
		os: 'linux',
		url: `${LATEST}/Sablier-linux-x64.deb`
	},
	{
		id: 'rpm',
		os: 'linux',
		url: `${LATEST}/Sablier-linux-x64.rpm`
	}
];

/** Guesses the visitor's desktop OS from the user agent, or null on a phone or an unknown system. */
export function detectOS(userAgent: string): DesktopOS | null {
	if (/android|iphone|ipad/i.test(userAgent)) return null;
	if (/mac os x|macintosh/i.test(userAgent)) return 'macos';
	if (/linux|x11/i.test(userAgent)) return 'linux';
	return null;
}

/** True on an Android phone or tablet, where the APK is the build to offer first. */
export function isAndroid(userAgent: string): boolean {
	return /android/i.test(userAgent);
}
