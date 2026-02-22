import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { useAuthContext } from '../auth/AuthContext';
import { createSession } from '../api/sessions';
import { Quiz } from '../types';
import { listMyQuizzes } from '../api/quizzes';

export default function Dashboard() {
  const { logout, email } = useAuthContext();
  const navigate = useNavigate();
  const [quizzes, setQuizzes] = useState<Quiz[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadQuizzes();
  }, []);

  const loadQuizzes = async () => {
    try {
      const data = await listMyQuizzes();
      setQuizzes(data || []);
    } catch (err) {
      console.error('Failed to load quizzes:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleHostGame = async (quizId: string) => {
    try {
      const session = await createSession(quizId);
      navigate(`/host/${session.sessionId}`);
    } catch (err) {
      console.error('Failed to create session:', err);
    }
  };

  return (
    <div className="min-h-screen bg-mesh">
      {/* Navbar */}
      <nav className="flex items-center justify-between px-6 py-4 border-b border-white/5">
        <Link to="/" className="text-xl font-display font-bold bg-gradient-to-r from-purple-400 to-blue-400 bg-clip-text text-transparent">
          KahootClone
        </Link>
        <div className="flex items-center gap-4">
          <span className="text-white/50 text-sm">{email}</span>
          <button onClick={logout} className="text-white/50 hover:text-white text-sm transition-colors">
            Sign Out
          </button>
        </div>
      </nav>

      <main className="max-w-5xl mx-auto px-6 py-8">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-3xl font-display font-bold">My Quizzes</h1>
          <Link to="/create-quiz" className="btn-primary">
            ✨ Create Quiz
          </Link>
        </div>

        {loading ? (
          <div className="text-center py-20 text-white/50">Loading quizzes...</div>
        ) : quizzes.length === 0 ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="glass-card p-12 text-center"
          >
            <div className="text-5xl mb-4">📝</div>
            <h3 className="text-xl font-display font-bold mb-2">No quizzes yet</h3>
            <p className="text-white/50 mb-6">Create your first quiz and host a live game!</p>
            <Link to="/create-quiz" className="btn-primary inline-block">Create Quiz</Link>
          </motion.div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {quizzes.map((quiz, i) => (
              <motion.div
                key={quiz.quizId}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.1 }}
                className="glass-card p-6 hover:border-purple-500/30 transition-colors"
              >
                <h3 className="text-lg font-display font-bold mb-1">{quiz.title}</h3>
                <p className="text-white/50 text-sm mb-3 line-clamp-2">{quiz.description}</p>
                <div className="flex items-center justify-between">
                  <span className="text-xs text-white/40">{quiz.questions.length} questions</span>
                  <button
                    onClick={() => handleHostGame(quiz.quizId)}
                    className="px-4 py-2 rounded-lg bg-kahoot-purple/20 text-purple-400 text-sm font-medium hover:bg-kahoot-purple/30 transition-colors"
                  >
                    🎮 Host Game
                  </button>
                </div>
              </motion.div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
