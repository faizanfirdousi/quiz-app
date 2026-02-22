import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { useQuizStore } from '../store/quizStore';
import { createQuiz } from '../api/quizzes';
import { Question, Option } from '../types';

const TIME_LIMITS = [5, 10, 20, 30, 60, 90, 120];
const ANSWER_COLORS = ['red', 'blue', 'yellow', 'green'];

function generateId() {
  return Math.random().toString(36).substr(2, 9);
}

function emptyQuestion(): Question {
  return {
    questionId: generateId(),
    text: '',
    options: [
      { id: generateId(), text: '' },
      { id: generateId(), text: '' },
      { id: generateId(), text: '' },
      { id: generateId(), text: '' },
    ],
    correctOptionId: '',
    timeLimitSeconds: 20,
    points: 1000,
  };
}

export default function CreateQuiz() {
  const navigate = useNavigate();
  const { title, description, questions, setTitle, setDescription, addQuestion, updateQuestion, removeQuestion, reset } = useQuizStore();
  const [currentIdx, setCurrentIdx] = useState(0);
  const [saving, setSaving] = useState(false);

  const handleAddQuestion = () => {
    addQuestion(emptyQuestion());
    setCurrentIdx(questions.length);
  };

  const handleSave = async () => {
    if (!title.trim()) return alert('Please enter a quiz title');
    if (questions.length === 0) return alert('Add at least one question');

    setSaving(true);
    try {
      await createQuiz({ title, description, questions });
      reset();
      navigate('/dashboard');
    } catch (err) {
      console.error('Failed to save quiz:', err);
      alert('Failed to save quiz');
    } finally {
      setSaving(false);
    }
  };

  const currentQuestion = questions[currentIdx] as Question | undefined;

  return (
    <div className="min-h-screen bg-mesh">
      {/* Header */}
      <nav className="flex items-center justify-between px-6 py-4 border-b border-white/5">
        <button onClick={() => navigate('/dashboard')} className="text-white/50 hover:text-white transition-colors">
          ← Back
        </button>
        <h1 className="text-lg font-display font-bold">Quiz Builder</h1>
        <button onClick={handleSave} disabled={saving} className="btn-primary text-sm disabled:opacity-50">
          {saving ? 'Saving...' : '💾 Save Quiz'}
        </button>
      </nav>

      <div className="max-w-5xl mx-auto px-6 py-6 grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Sidebar — question list */}
        <aside className="lg:col-span-1">
          <div className="glass-card p-4">
            <div className="space-y-3 mb-4">
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Quiz Title"
                className="input-field text-sm font-bold"
              />
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Description (optional)"
                className="input-field text-sm resize-none h-16"
              />
            </div>

            <h3 className="text-xs text-white/40 uppercase tracking-wider mb-2">Questions ({questions.length}/100)</h3>
            <div className="space-y-1 max-h-60 overflow-y-auto">
              {questions.map((q, i) => (
                <button
                  key={q.questionId}
                  onClick={() => setCurrentIdx(i)}
                  className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                    i === currentIdx ? 'bg-kahoot-purple/30 text-white' : 'text-white/60 hover:bg-white/5'
                  }`}
                >
                  {i + 1}. {q.text || 'Untitled'}
                </button>
              ))}
            </div>

            <button
              onClick={handleAddQuestion}
              disabled={questions.length >= 100}
              className="w-full mt-3 py-2 rounded-lg border border-dashed border-white/20 text-white/50 text-sm hover:bg-white/5 transition-colors disabled:opacity-30"
            >
              + Add Question
            </button>
          </div>
        </aside>

        {/* Main — question editor */}
        <main className="lg:col-span-3">
          <AnimatePresence mode="wait">
            {currentQuestion ? (
              <motion.div
                key={currentQuestion.questionId}
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
                className="glass-card p-6"
              >
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-lg font-display font-bold">Question {currentIdx + 1}</h2>
                  <button
                    onClick={() => {
                      removeQuestion(currentIdx);
                      setCurrentIdx(Math.max(0, currentIdx - 1));
                    }}
                    className="text-red-400/60 hover:text-red-400 text-sm transition-colors"
                  >
                    🗑 Delete
                  </button>
                </div>

                <textarea
                  value={currentQuestion.text}
                  onChange={(e) => {
                    const updated = { ...currentQuestion, text: e.target.value };
                    updateQuestion(currentIdx, updated);
                  }}
                  placeholder="Type your question here..."
                  className="input-field text-lg font-medium resize-none h-20 mb-6"
                />

                {/* Answer options */}
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-6">
                  {currentQuestion.options.map((option: Option, i: number) => (
                    <div
                      key={option.id}
                      className={`relative rounded-xl p-4 border-2 transition-all cursor-pointer ${
                        currentQuestion.correctOptionId === option.id
                          ? 'border-green-500 bg-green-500/10'
                          : 'border-white/10 hover:border-white/20'
                      }`}
                      onClick={() => {
                        const updated = { ...currentQuestion, correctOptionId: option.id };
                        updateQuestion(currentIdx, updated);
                      }}
                    >
                      <div className={`w-3 h-3 rounded-full mb-2 answer-btn ${ANSWER_COLORS[i]} p-0`} style={{ width: 12, height: 12 }} />
                      <input
                        type="text"
                        value={option.text}
                        onClick={(e) => e.stopPropagation()}
                        onChange={(e) => {
                          const newOptions = [...currentQuestion.options];
                          newOptions[i] = { ...option, text: e.target.value };
                          updateQuestion(currentIdx, { ...currentQuestion, options: newOptions });
                        }}
                        placeholder={`Option ${i + 1}`}
                        className="w-full bg-transparent border-none outline-none text-white placeholder-white/30"
                      />
                      {currentQuestion.correctOptionId === option.id && (
                        <span className="absolute top-2 right-2 text-green-400 text-xs">✓ Correct</span>
                      )}
                    </div>
                  ))}
                </div>

                {/* Settings */}
                <div className="flex flex-wrap gap-4">
                  <div>
                    <label className="block text-xs text-white/40 mb-1">Time Limit</label>
                    <select
                      value={currentQuestion.timeLimitSeconds}
                      onChange={(e) => {
                        updateQuestion(currentIdx, { ...currentQuestion, timeLimitSeconds: Number(e.target.value) });
                      }}
                      className="input-field text-sm w-28"
                    >
                      {TIME_LIMITS.map((t) => (
                        <option key={t} value={t}>{t}s</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-xs text-white/40 mb-1">Points</label>
                    <input
                      type="number"
                      value={currentQuestion.points}
                      onChange={(e) => {
                        updateQuestion(currentIdx, { ...currentQuestion, points: Number(e.target.value) });
                      }}
                      className="input-field text-sm w-28"
                      min={100}
                      max={5000}
                      step={100}
                    />
                  </div>
                </div>
              </motion.div>
            ) : (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="glass-card p-12 text-center"
              >
                <div className="text-4xl mb-4">📝</div>
                <h3 className="text-xl font-display font-bold mb-2">Start building your quiz</h3>
                <p className="text-white/50 mb-6">Add your first question to get started.</p>
                <button onClick={handleAddQuestion} className="btn-primary">+ Add Question</button>
              </motion.div>
            )}
          </AnimatePresence>
        </main>
      </div>
    </div>
  );
}
