import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { useAuthContext } from '../auth/AuthContext';

export default function Home() {
  const { isAuthenticated } = useAuthContext();

  return (
    <div className="min-h-screen bg-mesh flex flex-col">
      {/* Navbar */}
      <nav className="flex items-center justify-between px-6 py-4">
        <h1 className="text-2xl font-display font-bold bg-gradient-to-r from-purple-400 to-blue-400 bg-clip-text text-transparent">
          KahootClone
        </h1>
        <div className="flex gap-3">
          {isAuthenticated ? (
            <Link to="/dashboard" className="btn-primary text-sm">Dashboard</Link>
          ) : (
            <>
              <Link to="/login" className="px-4 py-2 rounded-xl text-white/70 hover:text-white transition-colors">Sign In</Link>
              <Link to="/register" className="btn-primary text-sm">Get Started</Link>
            </>
          )}
        </div>
      </nav>

      {/* Hero */}
      <main className="flex-1 flex items-center justify-center px-6">
        <div className="max-w-3xl text-center">
          <motion.h2
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            className="text-6xl md:text-7xl font-display font-extrabold mb-6 leading-tight"
          >
            Learn. Compete.{' '}
            <span className="bg-gradient-to-r from-kahoot-purple via-purple-400 to-kahoot-blue bg-clip-text text-transparent">
              Have Fun.
            </span>
          </motion.h2>

          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.2 }}
            className="text-xl text-white/60 mb-10 max-w-xl mx-auto"
          >
            Create interactive quizzes, host live games, and challenge players in real-time. 
            The ultimate quiz platform for classrooms, teams, and events.
          </motion.p>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.4 }}
            className="flex flex-col sm:flex-row gap-4 justify-center"
          >
            <Link
              to={isAuthenticated ? '/dashboard' : '/register'}
              className="btn-primary text-lg px-8 py-4"
            >
              🚀 Create a Quiz
            </Link>
            <Link
              to="/join"
              className="px-8 py-4 rounded-xl font-semibold text-lg border border-white/20 text-white hover:bg-white/5 transition-all duration-200"
            >
              🎮 Join a Game
            </Link>
          </motion.div>

          {/* Feature cards */}
          <motion.div
            initial={{ opacity: 0, y: 40 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.6 }}
            className="grid grid-cols-1 sm:grid-cols-3 gap-4 mt-16"
          >
            {[
              { emoji: '⚡', title: 'Real-Time', desc: 'Live WebSocket gameplay' },
              { emoji: '🏆', title: 'Leaderboards', desc: 'Instant scoring & ranks' },
              { emoji: '📱', title: 'Any Device', desc: 'Play on phone or desktop' },
            ].map((feature) => (
              <div key={feature.title} className="glass-card p-6 text-center">
                <div className="text-3xl mb-2">{feature.emoji}</div>
                <h3 className="font-display font-bold text-lg mb-1">{feature.title}</h3>
                <p className="text-white/50 text-sm">{feature.desc}</p>
              </div>
            ))}
          </motion.div>
        </div>
      </main>
    </div>
  );
}
