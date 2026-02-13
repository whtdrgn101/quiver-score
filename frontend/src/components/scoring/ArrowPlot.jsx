const RING_COLORS = [
  '#fbbf24', // X/10
  '#fbbf24', // 10
  '#fde68a', // 9
  '#ef4444', // 8
  '#ef4444', // 7
  '#3b82f6', // 6
  '#3b82f6', // 5
  '#1e1e1e', // 4
  '#1e1e1e', // 3
  '#f3f4f6', // 2
  '#f3f4f6', // 1
];

const RINGS = [
  { r: 100, fill: '#f3f4f6', stroke: '#d1d5db' },
  { r: 90, fill: '#f3f4f6', stroke: '#d1d5db' },
  { r: 80, fill: '#1e1e1e', stroke: '#374151' },
  { r: 70, fill: '#1e1e1e', stroke: '#374151' },
  { r: 60, fill: '#3b82f6', stroke: '#2563eb' },
  { r: 50, fill: '#3b82f6', stroke: '#2563eb' },
  { r: 40, fill: '#ef4444', stroke: '#dc2626' },
  { r: 30, fill: '#ef4444', stroke: '#dc2626' },
  { r: 20, fill: '#fde68a', stroke: '#d97706' },
  { r: 10, fill: '#fbbf24', stroke: '#d97706' },
];

export default function ArrowPlot({ ends }) {
  const arrows = ends.flatMap((e) => e.arrows).filter((a) => a.x_pos != null && a.y_pos != null);
  if (arrows.length === 0) return null;

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
      <h3 className="text-sm font-semibold text-gray-500 dark:text-gray-400 mb-3">Arrow Placement</h3>
      <div className="flex justify-center">
        <svg viewBox="-110 -110 220 220" className="w-full max-w-[280px]">
          {RINGS.map((ring, i) => (
            <circle key={i} cx={0} cy={0} r={ring.r} fill={ring.fill} stroke={ring.stroke} strokeWidth={0.5} />
          ))}
          {/* Cross-hairs */}
          <line x1={-105} y1={0} x2={105} y2={0} stroke="#9ca3af" strokeWidth={0.3} />
          <line x1={0} y1={-105} x2={0} y2={105} stroke="#9ca3af" strokeWidth={0.3} />
          {/* Arrows */}
          {arrows.map((a) => (
            <circle
              key={a.id}
              cx={a.x_pos}
              cy={-a.y_pos}
              r={3}
              fill="#10b981"
              stroke="#065f46"
              strokeWidth={0.8}
              opacity={0.85}
            >
              <title>{a.score_value}</title>
            </circle>
          ))}
        </svg>
      </div>
    </div>
  );
}
