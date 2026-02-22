import { motion } from 'framer-motion';
import { Option } from '../types';

interface QuestionCardProps {
  text: string;
  options: Option[];
  questionIndex: number;
  totalQuestions: number;
  points: number;
}

export default function QuestionCard({ text, options, questionIndex, totalQuestions, points }: QuestionCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="glass-card p-6"
    >
      <div className="flex items-center justify-between mb-3">
        <span className="text-xs text-white/40 font-display">
          Question {questionIndex + 1} / {totalQuestions}
        </span>
        <span className="text-xs text-purple-400 font-display font-bold">
          {points} pts
        </span>
      </div>

      <h3 className="text-xl font-display font-bold mb-4">{text}</h3>

      <div className="grid grid-cols-2 gap-2">
        {options.map((option, i) => (
          <div
            key={option.id}
            className="px-3 py-2 rounded-lg bg-white/5 text-sm text-white/70"
          >
            <span className="text-white/30 mr-2">{['A', 'B', 'C', 'D'][i]}.</span>
            {option.text}
          </div>
        ))}
      </div>
    </motion.div>
  );
}
