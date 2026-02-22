import { motion } from 'framer-motion';

interface AnswerButtonProps {
  label: string;
  shape: string;
  color: 'red' | 'blue' | 'yellow' | 'green';
  onClick: () => void;
  disabled: boolean;
  index: number;
}

export default function AnswerButton({ label, shape, color, onClick, disabled, index }: AnswerButtonProps) {
  return (
    <motion.button
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.1 }}
      whileHover={!disabled ? { scale: 1.02 } : undefined}
      whileTap={!disabled ? { scale: 0.98 } : undefined}
      onClick={onClick}
      disabled={disabled}
      className={`answer-btn ${color}`}
    >
      <span className="mr-3 opacity-60">{shape}</span>
      {label}
    </motion.button>
  );
}
