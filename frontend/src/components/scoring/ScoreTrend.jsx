import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, Tooltip, CartesianGrid } from 'recharts';

export default function ScoreTrend({ data }) {
  if (!data || data.length < 2) return null;

  const chartData = [...data].reverse().map((t) => ({
    date: new Date(t.date).toLocaleDateString(undefined, { month: 'short', day: 'numeric' }),
    percentage: t.max_score > 0 ? +((t.score / t.max_score) * 100).toFixed(1) : 0,
    score: t.score,
    round: t.template_name,
  }));

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
      <h3 className="text-sm font-semibold text-gray-500 dark:text-gray-400 mb-3">Score Trend</h3>
      <ResponsiveContainer width="100%" height={200}>
        <LineChart data={chartData}>
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" opacity={0.3} />
          <XAxis dataKey="date" tick={{ fontSize: 12 }} stroke="#9ca3af" />
          <YAxis domain={[0, 100]} tick={{ fontSize: 12 }} stroke="#9ca3af" unit="%" />
          <Tooltip
            contentStyle={{ backgroundColor: '#1f2937', border: 'none', borderRadius: '8px', color: '#fff' }}
            formatter={(value, name, props) => [`${props.payload.score} (${value}%)`, props.payload.round]}
            labelStyle={{ color: '#9ca3af' }}
          />
          <Line type="monotone" dataKey="percentage" stroke="#10b981" strokeWidth={2} dot={{ fill: '#10b981', r: 4 }} activeDot={{ r: 6 }} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
