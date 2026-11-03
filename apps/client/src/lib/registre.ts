/** Where shared spaces are managed. Sablier only mirrors them. */
const REGISTRE_SPACES_URL = 'https://sso.facile.studio/spaces';

/**
 * Returns the Registre page for a space, or the list of spaces when it is not linked yet.
 */
export function registreSpaceUrl(registreSpaceId?: string): string {
	return registreSpaceId ? `${REGISTRE_SPACES_URL}/${registreSpaceId}` : REGISTRE_SPACES_URL;
}
