import { useEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { useSiteStore } from './store/site';
import BottomNav from './components/features/BottomNav';
import LinkNoticeSheet from './components/features/LinkNoticeSheet';
import FlyerViewer from './components/features/FlyerViewer';
import CookieBanner from './components/features/CookieBanner';
import CalendarPage from './pages/CalendarPage';
import AddEventPage from './pages/AddEventPage';
import FilterPage from './pages/FilterPage';
import FeedsPage from './pages/FeedsPage';
import MorePage from './pages/MorePage';
import PopBookerPage from './pages/PopBookerPage';
import AboutPage from './pages/AboutPage';
import AdminLoginPage from './pages/AdminLoginPage';
import AdminDashboardPage from './pages/AdminDashboardPage';
import AdminAddEventPage from './pages/AdminAddEventPage';
import VerifyEventPage from './pages/VerifyEventPage';
import PrivacyPolicy from './pages/PrivacyPolicy';

function App() {
  const loadSite = useSiteStore((state) => state.load);
  useEffect(() => {
    void loadSite();
  }, [loadSite]);

  return (
    <BrowserRouter>
      <div className="mx-auto flex h-dvh w-full max-w-[480px] flex-col bg-bg text-text sm:border-x sm:border-border">
        <main className="flex min-h-0 flex-1 flex-col">
          <Routes>
            <Route path="/" element={<CalendarPage />} />
            <Route path="/calendar" element={<CalendarPage />} />
            <Route path="/add-event" element={<AddEventPage />} />
            <Route path="/filter" element={<FilterPage />} />
            <Route path="/feeds" element={<FeedsPage />} />
            <Route path="/more" element={<MorePage />} />
            <Route path="/about" element={<AboutPage />} />
            <Route path="/pop-booker" element={<PopBookerPage />} />
            <Route path="/admin/login" element={<AdminLoginPage />} />
            <Route path="/admin" element={<AdminDashboardPage />} />
            <Route path="/admin/add-event" element={<AdminAddEventPage />} />
            <Route path="/verify-event" element={<VerifyEventPage />} />
            <Route path="/privacy-policy" element={<PrivacyPolicy />} />
          </Routes>
        </main>
        <CookieBanner />
        <BottomNav />
        <LinkNoticeSheet />
        <FlyerViewer />
      </div>
    </BrowserRouter>
  );
}

export default App;
