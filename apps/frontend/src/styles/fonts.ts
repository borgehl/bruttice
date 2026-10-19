import '@fontsource-variable/public-sans';

let neonFonts: Promise<unknown> | undefined;

// Neon's faces are only fetched once the theme is first used, so the default theme never pays for them.
export function loadNeonFonts(): Promise<unknown> {
  neonFonts ??= Promise.all([
    import('@fontsource/chakra-petch/600.css'),
    import('@fontsource/chakra-petch/700.css'),
    import('@fontsource-variable/jetbrains-mono'),
    import('@fontsource-variable/inter-tight'),
  ]);
  return neonFonts;
}
