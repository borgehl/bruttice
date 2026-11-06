import { useEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { useSiteStore } from './store/site';
import { usePagesStore } from './store/pages';
import BottomNav from './components/features/BottomNav';
import DesktopHeader from './components/features/DesktopHeader';
import DesktopFooter from './components/features/DesktopFooter';
import LinkNoticeSheet from './components/features/LinkNoticeSheet';
import FlyerViewer from './components/features/FlyerViewer';
import CookieBanner from './components/features/CookieBanner';
import CalendarPage from './pages/CalendarPage';
import AddEventPage from './pages/AddEventPage';
import FilterPage from './pages/FilterPage';
import FeedsPage from './pages/FeedsPage';
import MorePage from './pages/MorePage';
import ContentPage from './pages/ContentPage';
import AboutPage from './pages/AboutPage';
import AdminLoginPage from './pages/AdminLoginPage';
import AdminDashboardPage from './pages/AdminDashboardPage';
import AdminAddEventPage from './pages/AdminAddEventPage';
import VerifyEventPage from './pages/VerifyEventPage';
import PrivacyPolicy from './pages/PrivacyPolicy';
import ContactPage from './pages/ContactPage';
import AdminSettingsPage from './pages/AdminSettingsPage';
import AdminConfigPage from './pages/AdminConfigPage';

function App() {
  const loadSite = useSiteStore((state) => state.load);
  const siteName = useSiteStore((state) => state.site.name);
  const siteLoaded = useSiteStore((state) => state.loaded);
  const loadPages = usePagesStore((state) => state.load);
  useEffect(() => {
    void loadSite();
    void loadPages();
  }, [loadSite, loadPages]);

  // The served page already has the name when the image is given SITE_NAME. This covers a server
  // without it, and a name changed in the config since. Until the config has loaded the store only
  // holds its default, which must not replace the title the page came with.
  useEffect(() => {
    if (siteLoaded) {
      document.title = siteName;
    }
  }, [siteLoaded, siteName]);

  return (
    <BrowserRouter>
      <div className="flex h-dvh flex-col bg-bg text-text">
        <DesktopHeader />
        {/* a phone-width column below tablet size, a wider one from there up */}
        <div className="mx-auto flex min-h-0 w-full max-w-[480px] flex-1 flex-col sm:border-x sm:border-border md:max-w-4xl">
        <main className="flex min-h-0 flex-1 flex-col">
          <Routes>
            <Route path="/" element={<CalendarPage />} />
            <Route path="/calendar" element={<CalendarPage />} />
            <Route path="/add-event" element={<AddEventPage />} />
            <Route path="/filter" element={<FilterPage />} />
            <Route path="/feeds" element={<FeedsPage />} />
            <Route path="/more" element={<MorePage />} />
            <Route path="/about" element={<AboutPage />} />
            <Route path="/contact" element={<ContactPage />} />
            <Route path="/p/:slug" element={<ContentPage />} />
            <Route path="/admin/login" element={<AdminLoginPage />} />
            <Route path="/admin" element={<AdminDashboardPage />} />
            <Route path="/admin/add-event" element={<AdminAddEventPage />} />
            <Route path="/admin/settings" element={<AdminSettingsPage />} />
            <Route path="/admin/config" element={<AdminConfigPage />} />
            <Route path="/verify-event" element={<VerifyEventPage />} />
            <Route path="/privacy-policy" element={<PrivacyPolicy />} />
          </Routes>
        </main>
        <CookieBanner />
        </div>
        <DesktopFooter />
        <BottomNav />
        <LinkNoticeSheet />
        <FlyerViewer />
      </div>
    </BrowserRouter>
  );
}

export default App;
