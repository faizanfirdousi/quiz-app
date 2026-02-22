import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuthContext } from './auth/AuthContext';
import Home from './pages/Home';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import CreateQuiz from './pages/CreateQuiz';
import HostLobby from './pages/HostLobby';
import PlayerJoin from './pages/PlayerJoin';
import PlayerLobby from './pages/PlayerLobby';
import Question from './pages/Question';
import AnswerResult from './pages/AnswerResult';
import Leaderboard from './pages/Leaderboard';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuthContext();

  if (isLoading) {
    return (
      <div className="min-h-screen bg-mesh flex items-center justify-center">
        <div className="animate-pulse-slow text-2xl font-display text-white/60">Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

export default function App() {
  return (
    <div className="min-h-screen bg-mesh">
      <Routes>
        {/* Public routes */}
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/join" element={<PlayerJoin />} />

        {/* Protected routes */}
        <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
        <Route path="/create-quiz" element={<ProtectedRoute><CreateQuiz /></ProtectedRoute>} />
        <Route path="/host/:sessionId" element={<ProtectedRoute><HostLobby /></ProtectedRoute>} />

        {/* Game routes (can be accessed by players with PIN) */}
        <Route path="/lobby/:sessionId" element={<PlayerLobby />} />
        <Route path="/play/:sessionId" element={<Question />} />
        <Route path="/result/:sessionId" element={<AnswerResult />} />
        <Route path="/leaderboard/:sessionId" element={<Leaderboard />} />
      </Routes>
    </div>
  );
}
