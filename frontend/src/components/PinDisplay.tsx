import { motion } from 'framer-motion';

interface PinDisplayProps {
  pin: string;
}

export default function PinDisplay({ pin }: PinDisplayProps) {
  return (
    <div className="glass-card p-8 inline-block">
      <p className="text-white/40 text-sm mb-2 uppercase tracking-wider">Game PIN</p>
      <motion.div
        initial={{ scale: 0.8 }}
        animate={{ scale: 1 }}
        className="pin-display"
      >
        {pin.split('').map((char, i) => (
          <motion.span
            key={i}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.1 }}
          >
            {char}
          </motion.span>
        ))}
      </motion.div>
    </div>
  );
}
