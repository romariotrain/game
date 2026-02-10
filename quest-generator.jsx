import { useState, useEffect, useRef } from "react";

const USER_PROFILE = `
Имя: Рома, 22 года
Образование: менеджмент (получено), сейчас учится на DevOps в универе
Карьера: ищет работу Go-разработчиком, делает свой проект (Solo Leveling ToDo игра)
Оборудование дома: беговая дорожка, коврик для йоги
Хобби: шахматы, Valorant (хочет прокачаться), аниме (Solo Leveling)
Проблемы: сбитый режим сна, нужна мотивация заниматься здоровьем
Цели: найти работу Go-разработчиком, наладить режим, улучшиться в Valorant, заняться здоровьем
Технологии: Go, DevOps, PostgreSQL, SQLite, Fyne
`;

const SYSTEM_PROMPT = `Ты — система генерации заданий для Solo Leveling Life RPG приложения.

Профиль пользователя:
${USER_PROFILE}

Система заданий:
- Статы: STR (Сила — физическая активность), AGI (Ловкость — реакция, Valorant, ловкость), INT (Интеллект — учёба, программирование, шахматы), STA (Выносливость — здоровье, режим, дорожка)
- Ранги: E (очень лёгкое, 20 EXP), D (лёгкое, 40 EXP), C (среднее, 70 EXP), B (сложное, 120 EXP), A (очень сложное, 200 EXP), S (эпическое, 350 EXP)
- Одно задание может качать несколько статов (указывай массив)

Твоя задача:
1. Сначала задай 2-3 уточняющих вопроса о дне/неделе пользователя
2. После ответа сгенерируй 4-6 заданий на день в формате JSON
3. Учитывай реальные обстоятельства (усталость, время, цели)
4. Делай задания конкретными и выполнимыми

ВАЖНО: Когда генерируешь задания, ВСЕГДА заканчивай сообщение блоком JSON в точно таком формате:
\`\`\`json
[
  {
    "name": "Название задания",
    "description": "Конкретное описание что нужно сделать",
    "rank": "D",
    "stats": ["STA"],
    "is_daily": false
  }
]
\`\`\`

Пиши по-русски. Будь как наставник — понимающий но мотивирующий.`;

const WEEK_SYSTEM_PROMPT = `Ты — система генерации заданий для Solo Leveling Life RPG приложения. Сейчас начало недели.

Профиль пользователя:
${USER_PROFILE}

Система заданий:
- Статы: STR (Сила — физическая активность), AGI (Ловкость — реакция, Valorant, ловкость), INT (Интеллект — учёба, программирование, шахматы), STA (Выносливость — здоровье, режим, дорожка)
- Ранги: E (очень лёгкое, 20 EXP), D (лёгкое, 40 EXP), C (среднее, 70 EXP), B (сложное, 120 EXP), A (очень сложное, 200 EXP), S (эпическое, 350 EXP)

Твоя задача:
1. Задай 4-5 вопросов о прошедшей неделе и планах на текущую
2. После ответа сгенерируй набор ЕЖЕДНЕВНЫХ заданий (is_daily: true) на неделю — базовые привычки которые повторяются каждый день
3. Плюс 3-5 недельных задач (is_daily: false) — важные цели на неделю

ВАЖНО: Когда генерируешь задания, ВСЕГДА заканчивай блоком JSON:
\`\`\`json
[...]
\`\`\`

Пиши по-русски. Будь как наставник.`;

const RANK_COLORS = {
    E: "#6b7280",
    D: "#3b82f6",
    C: "#10b981",
    B: "#f59e0b",
    A: "#ef4444",
    S: "#8b5cf6",
};

const STAT_ICONS = {
    STR: "⚔️",
    AGI: "⚡",
    INT: "🧠",
    STA: "❤️",
};

export default function QuestGenerator() {
    const [mode, setMode] = useState(null); // 'day' | 'week'
    const [messages, setMessages] = useState([]);
    const [input, setInput] = useState("");
    const [loading, setLoading] = useState(false);
    const [quests, setQuests] = useState(null);
    const [copied, setCopied] = useState(false);
    const [phase, setPhase] = useState("questions"); // 'questions' | 'generated'
    const messagesEndRef = useRef(null);
    const inputRef = useRef(null);

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages]);

    const startMode = async (selectedMode) => {
        setMode(selectedMode);
        setPhase("questions");
        setQuests(null);

        const greeting = selectedMode === "day"
            ? "Привет, Рома! Давай составим задания на сегодня. Сначала пару вопросов..."
            : "Привет, Рома! Начало новой недели — самое время всё спланировать. Несколько вопросов...";

        const initMessages = [{ role: "user", content: greeting }];
        setMessages([{ role: "assistant", content: "..." }]);
        setLoading(true);

        try {
            const systemPrompt = selectedMode === "day" ? SYSTEM_PROMPT : WEEK_SYSTEM_PROMPT;
            const response = await fetch("https://api.anthropic.com/v1/messages", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    model: "claude-sonnet-4-5-20250929",
                    max_tokens: 1000,
                    system: systemPrompt,
                    messages: initMessages,
                }),
            });
            const data = await response.json();
            const text = data.content?.[0]?.text || "Ошибка соединения";
            setMessages([{ role: "assistant", content: text }]);
        } catch (e) {
            setMessages([{ role: "assistant", content: "Ошибка соединения с сервером." }]);
        }
        setLoading(false);
    };

    const extractJSON = (text) => {
        const match = text.match(/```json\n([\s\S]*?)\n```/);
        if (match) {
            try {
                return JSON.parse(match[1]);
            } catch {}
        }
        return null;
    };

    const sendMessage = async () => {
        if (!input.trim() || loading) return;

        const userMessage = input.trim();
        setInput("");

        const newMessages = [...messages, { role: "user", content: userMessage }];
        setMessages(newMessages);
        setLoading(true);

        const assistantPlaceholder = [...newMessages, { role: "assistant", content: "..." }];
        setMessages(assistantPlaceholder);

        try {
            const systemPrompt = mode === "day" ? SYSTEM_PROMPT : WEEK_SYSTEM_PROMPT;
            const response = await fetch("https://api.anthropic.com/v1/messages", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    model: "claude-sonnet-4-5-20250929",
                    max_tokens: 1500,
                    system: systemPrompt,
                    messages: newMessages,
                }),
            });
            const data = await response.json();
            const text = data.content?.[0]?.text || "Ошибка";

            const finalMessages = [...newMessages, { role: "assistant", content: text }];
            setMessages(finalMessages);

            const extracted = extractJSON(text);
            if (extracted) {
                setQuests(extracted);
                setPhase("generated");
            }
        } catch (e) {
            const errorMessages = [...newMessages, { role: "assistant", content: "Ошибка соединения." }];
            setMessages(errorMessages);
        }
        setLoading(false);
    };

    const copyJSON = () => {
        if (!quests) return;
        navigator.clipboard.writeText(JSON.stringify(quests, null, 2));
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    const reset = () => {
        setMode(null);
        setMessages([]);
        setInput("");
        setQuests(null);
        setPhase("questions");
    };

    const renderMessage = (msg, idx) => {
        const isUser = msg.role === "user";
        const isLoading = msg.content === "...";

        // Remove JSON block from display
        const displayContent = msg.content.replace(/```json\n[\s\S]*?\n```/g, "").trim();

        return (
            <div key={idx} className={`flex ${isUser ? "justify-end" : "justify-start"} mb-4`}>
                {!isUser && (
                    <div className="w-8 h-8 rounded-full bg-blue-600 flex items-center justify-center text-xs font-bold mr-2 mt-1 flex-shrink-0">
                        SL
                    </div>
                )}
                <div
                    className={`max-w-xs lg:max-w-md px-4 py-3 rounded-2xl text-sm leading-relaxed ${
                        isUser
                            ? "bg-blue-600 text-white rounded-tr-sm"
                            : "bg-gray-800 text-gray-100 rounded-tl-sm border border-gray-700"
                    }`}
                >
                    {isLoading ? (
                        <div className="flex gap-1 items-center py-1">
                            <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: "0ms" }} />
                            <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: "150ms" }} />
                            <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: "300ms" }} />
                        </div>
                    ) : (
                        <span style={{ whiteSpace: "pre-wrap" }}>{displayContent}</span>
                    )}
                </div>
                {isUser && (
                    <div className="w-8 h-8 rounded-full bg-gray-600 flex items-center justify-center text-xs font-bold ml-2 mt-1 flex-shrink-0">
                        Р
                    </div>
                )}
            </div>
        );
    };

    // Landing screen
    if (!mode) {
        return (
            <div className="min-h-screen bg-gray-950 text-white flex flex-col items-center justify-center p-6">
                <div className="text-center mb-12">
                    <div className="text-6xl mb-4">⚔️</div>
                    <h1 className="text-3xl font-bold mb-2 tracking-tight">
                        Quest Generator
                    </h1>
                    <p className="text-gray-400 text-sm">Solo Leveling · Рома · Lvl ???</p>
                </div>

                <div className="flex flex-col gap-4 w-full max-w-sm">
                    <button
                        onClick={() => startMode("day")}
                        className="bg-blue-600 hover:bg-blue-500 transition-colors rounded-xl p-5 text-left group"
                    >
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-2xl">🌅</span>
                            <span className="text-xs text-blue-300 opacity-0 group-hover:opacity-100 transition-opacity">→</span>
                        </div>
                        <div className="font-bold text-lg">Задания на день</div>
                        <div className="text-blue-200 text-sm mt-1">3-5 заданий с учётом твоего дня</div>
                    </button>

                    <button
                        onClick={() => startMode("week")}
                        className="bg-gray-800 hover:bg-gray-700 border border-gray-700 transition-colors rounded-xl p-5 text-left group"
                    >
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-2xl">📅</span>
                            <span className="text-xs text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity">→</span>
                        </div>
                        <div className="font-bold text-lg">Планирование недели</div>
                        <div className="text-gray-400 text-sm mt-1">Ежедневные привычки + цели на неделю</div>
                    </button>
                </div>

                <div className="mt-8 text-xs text-gray-600 text-center">
                    Профиль: Go-разработчик · Valorant · Дорожка + йога
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gray-950 text-white flex flex-col" style={{ maxHeight: "100vh" }}>
            {/* Header */}
            <div className="flex items-center justify-between px-4 py-3 border-b border-gray-800 bg-gray-900 flex-shrink-0">
                <button onClick={reset} className="text-gray-400 hover:text-white transition-colors text-sm flex items-center gap-1">
                    ← Назад
                </button>
                <span className="text-sm font-medium text-gray-300">
          {mode === "day" ? "🌅 На день" : "📅 На неделю"}
        </span>
                <div className="w-16" />
            </div>

            {/* Messages */}
            <div className="flex-1 overflow-y-auto px-4 py-4">
                {messages.map((msg, idx) => renderMessage(msg, idx))}
                <div ref={messagesEndRef} />
            </div>

            {/* Generated quests panel */}
            {quests && phase === "generated" && (
                <div className="flex-shrink-0 border-t border-gray-800 bg-gray-900 px-4 py-4 max-h-64 overflow-y-auto">
                    <div className="flex items-center justify-between mb-3">
                        <span className="text-sm font-bold text-green-400">✓ Сгенерировано {quests.length} заданий</span>
                        <button
                            onClick={copyJSON}
                            className={`text-xs px-3 py-1.5 rounded-lg font-medium transition-all ${
                                copied
                                    ? "bg-green-600 text-white"
                                    : "bg-blue-600 hover:bg-blue-500 text-white"
                            }`}
                        >
                            {copied ? "✓ Скопировано!" : "Копировать JSON"}
                        </button>
                    </div>

                    <div className="flex flex-col gap-2">
                        {quests.map((q, i) => (
                            <div key={i} className="bg-gray-800 rounded-lg px-3 py-2 flex items-start gap-3 border border-gray-700">
                <span
                    className="text-xs font-bold px-1.5 py-0.5 rounded mt-0.5 flex-shrink-0"
                    style={{
                        backgroundColor: RANK_COLORS[q.rank] + "33",
                        color: RANK_COLORS[q.rank],
                        border: `1px solid ${RANK_COLORS[q.rank]}55`,
                    }}
                >
                  {q.rank}
                </span>
                                <div className="flex-1 min-w-0">
                                    <div className="text-sm font-medium text-white truncate">{q.name}</div>
                                    <div className="text-xs text-gray-400 mt-0.5 leading-relaxed">{q.description}</div>
                                </div>
                                <div className="flex gap-1 flex-shrink-0">
                                    {(q.stats || [q.stat]).filter(Boolean).map((s) => (
                                        <span key={s} className="text-sm" title={s}>{STAT_ICONS[s] || s}</span>
                                    ))}
                                    {q.is_daily && <span className="text-xs" title="Ежедневное">🔄</span>}
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            )}

            {/* Input */}
            <div className="flex-shrink-0 px-4 py-3 border-t border-gray-800 bg-gray-900">
                <div className="flex gap-2">
                    <input
                        ref={inputRef}
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        onKeyDown={(e) => e.key === "Enter" && !e.shiftKey && sendMessage()}
                        placeholder="Напиши ответ..."
                        disabled={loading}
                        className="flex-1 bg-gray-800 border border-gray-700 rounded-xl px-4 py-2.5 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 transition-colors disabled:opacity-50"
                    />
                    <button
                        onClick={sendMessage}
                        disabled={loading || !input.trim()}
                        className="bg-blue-600 hover:bg-blue-500 disabled:opacity-40 disabled:cursor-not-allowed transition-colors rounded-xl px-4 py-2.5 text-sm font-medium"
                    >
                        →
                    </button>
                </div>
            </div>
        </div>
    );
}
