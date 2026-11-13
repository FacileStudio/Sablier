import type { CliCommand, DownloadEntry } from '@facile/muse';
import { t } from '$lib/i18n/index.svelte';

const RELEASES = 'https://github.com/FacileStudio/Sablier/releases';

const latest = (asset: string) => `${RELEASES}/latest/download/${asset}`;

export const APK_URL = latest('Sablier-android-universal.apk');
export const APK_SHA256_URL = `${APK_URL}.sha256`;
export const RELEASES_URL = RELEASES;
export const DESKTOP_CHECKSUMS_URL = latest('desktop-checksums.txt');

export const DOWNLOADS: DownloadEntry[] = [
	{ platform: 'macos', href: latest('Sablier-macos-universal.dmg'), label: '.dmg', arch: 'universal' },
	{ platform: 'windows', href: latest('Sablier-windows-x64-setup.exe'), label: '.exe', arch: 'x64' },
	{ platform: 'linux', href: latest('Sablier-linux-x64.deb'), label: '.deb', arch: 'x64' },
	{ platform: 'linux', href: latest('Sablier-linux-x64.rpm'), label: '.rpm', arch: 'x64' },
	{ platform: 'android', href: APK_URL, label: 'APK', arch: 'universal', checksumHref: APK_SHA256_URL }
];

export const INSTALL_METHODS: CliCommand[] = [
	{
		get label() {
			return t('downloads.install.script');
		},
		command: 'curl -fsSL https://raw.githubusercontent.com/FacileStudio/sablier-cli/main/install.sh | bash'
	},
	{
		get label() {
			return t('downloads.install.facile');
		},
		command: 'facile install sablier'
	},
	{
		get label() {
			return t('downloads.install.brew');
		},
		command: 'brew install FacileStudio/tap/sablier'
	}
];
