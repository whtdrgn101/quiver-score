import { ResponsiveContainer, BarChart, Bar, XAxis, YAxis, Tooltip, CartesianGrid } from 'recharts';

export default function EndBarChart({ ends, maxPerEnd }) {
  if (!ends || ends.length === 0) return null;

  const chartData = ends.map((e) => ({
    end: `E${e.end_number}`,
    score: e.end_total,
  }));

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
      <h3 className="text-sm font-semibold text-gray-500 dark:text-gray-400 mb-3">Score per End</h3>
      <ResponsiveContainer width="100%" height={200}>
        <BarChart data={chartData}>
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" opacity={0.3} />
          <XAxis dataKey="end" tick={{ fontSize: 12 }} stroke="#9ca3af" />
          <YAxis domain={[0, maxPerEnd || 'auto']} tick={{ fontSize: 12 }} stroke="#9ca3af" />
          <Tooltip
            contentStyle={{ backgroundColor: '#1f2937', border: 'none', borderRadius: '8px', color: '#fff' }}
            labelStyle={{ color: '#9ca3af' }}
          />
          <Bar dataKey="score" fill="#10b981" radius={[4, 4, 0, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
