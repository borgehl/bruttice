import Segmented from '../../ui/Segmented';
import { useThemeStore, type Theme } from '../../../store/theme';

const options: { value: Theme; label: string }[] = [
  { value: 'info', label: 'Info' },
  { value: 'neon', label: 'Neon' },
];

// The theme switch on its own, for the desktop header. On a phone it is a row in More.
export default function ThemeSwitchCompact() {
  const { theme, setTheme } = useThemeStore();
  return <Segmented label="Theme" value={theme} options={options} onChange={setTheme} />;
}
