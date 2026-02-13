export default function Spinner({ text = 'Loading...' }) {
  return (
    <div className="flex items-center justify-center gap-2 py-8">
      <div className="w-5 h-5 border-2 border-emerald-600 border-t-transparent rounded-full animate-spin" />
      <span className="text-gray-500 dark:text-gray-400 text-sm">{text}</span>
    </div>
  );
}
