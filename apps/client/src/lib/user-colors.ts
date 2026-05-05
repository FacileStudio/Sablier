export const USER_COLORS = ['#AD9EF0', '#F09ED6', '#EE7E89', '#EEB47E', '#A9EE7E', '#7EEEDB'] as const;

export type UserColor = (typeof USER_COLORS)[number];

export function normalizeUserColor(color: string | null | undefined): UserColor {
	const normalized = color?.trim().toUpperCase();
	if (!normalized) {
		return USER_COLORS[0];
	}
	const withHash = normalized.startsWith('#') ? normalized : `#${normalized}`;
	return (USER_COLORS as readonly string[]).includes(withHash) ? (withHash as UserColor) : USER_COLORS[0];
}
