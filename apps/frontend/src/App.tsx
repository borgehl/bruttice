import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Header from './components/features/Header';
import Footer from './components/features/Footer';
import CookieBanner from './components/features/CookieBanner';
import CalendarPage from './pages/CalendarPage';
import AddEventPage from './pages/AddEventPage';
import FilterPage from './pages/FilterPage';
import FeedsPage from './pages/FeedsPage';
import AboutPage from './pages/AboutPage';
import AdminLoginPage from './pages/AdminLoginPage';
import AdminDashboardPage from './pages/AdminDashboardPage';
import AdminAddEventPage from './pages/AdminAddEventPage';
import VerifyEventPage from './pages/VerifyEventPage';
import PrivacyPolicy from './pages/PrivacyPolicy';

function App() {
  return (
    <BrowserRouter>
      <div className="min-h-screen flex flex-col bg-background text-text">
        <Header />
        <main className="flex-1 max-w-4xl w-full mx-auto px-4 py-8">
          <Routes>
            <Route path="/" element={<CalendarPage />} />
            <Route path="/calendar" element={<CalendarPage />} />
            <Route path="/add-event" element={<AddEventPage />} />
            <Route path="/filter" element={<FilterPage />} />
            <Route path="/feeds" element={<FeedsPage />} />
            <Route path="/about" element={<AboutPage />} />
            <Route path="/admin/login" element={<AdminLoginPage />} />
            <Route path="/admin" element={<AdminDashboardPage />} />
            <Route path="/admin/add-event" element={<AdminAddEventPage />} />
            <Route path="/verify-event" element={<VerifyEventPage />} />
            <Route path="/privacy-policy" element={<PrivacyPolicy />} />
          </Routes>
        </main>
        <Footer />
        <CookieBanner />
      </div>
    </BrowserRouter>
  );
}

export default App;
