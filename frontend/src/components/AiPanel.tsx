import { useState } from 'react'
import { Sparkles, MessageSquareText, Smile, Frown, Meh, Loader2, X, Copy, Check } from 'lucide-react'
import { aiService } from '../services/ai'

interface Props {
  conversationId: string
  onInsertReply: (text: string) => void
  onClose: () => void
}

/**
 * AiPanel — right-side panel of Claude-powered actions for the open
 * conversation: suggest reply, summarize thread, sentiment tag.
 */
export default function AiPanel({ conversationId, onInsertReply, onClose }: Props) {
  const [tab, setTab] = useState<'suggest' | 'summary' | 'sentiment'>('suggest')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [suggestion, setSuggestion] = useState('')
  const [summary, setSummary] = useState('')
  const [sentiment, setSentiment] = useState<{ score: string; reason: string } | null>(null)
  const [devMode, setDevMode] = useState(false)
  const [copied, setCopied] = useState(false)

  const run = async () => {
    setLoading(true); setError('')
    try {
      if (tab === 'suggest') {
        const r = await aiService.suggestReply(conversationId)
        setSuggestion(r.suggestion); setDevMode(r.dev_mode)
      } else if (tab === 'summary') {
        const r = await aiService.summarize(conversationId)
        setSummary(r.summary); setDevMode(r.dev_mode)
      } else {
        const r = await aiService.sentiment(conversationId)
        setSentiment({ score: r.score, reason: r.reason }); setDevMode(r.dev_mode)
      }
    } catch (err: any) {
      setError(err.response?.data?.detail || 'AI request failed')
    } finally {
      setLoading(false)
    }
  }

  const doCopy = (text: string) => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    })
  }

  return (
    <aside className="w-80 border-l border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex flex-col">
      <div className="px-4 py-3 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
        <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
          <Sparkles size={16} className="text-primary-600" />
          AI Assist
        </h3>
        <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"><X size={16} /></button>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-gray-100 dark:border-gray-700 text-xs">
        <TabBtn active={tab === 'suggest'} onClick={() => setTab('suggest')}><Sparkles size={12} /> Reply</TabBtn>
        <TabBtn active={tab === 'summary'} onClick={() => setTab('summary')}><MessageSquareText size={12} /> Summary</TabBtn>
        <TabBtn active={tab === 'sentiment'} onClick={() => setTab('sentiment')}><Smile size={12} /> Sentiment</TabBtn>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {devMode && (
          <div className="rounded-lg bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 px-2.5 py-1.5 text-[10px] text-amber-800 dark:text-amber-300">
            Dev mode — set ANTHROPIC_API_KEY for real Claude responses.
          </div>
        )}

        <button
          onClick={run}
          disabled={loading}
          className="w-full flex items-center justify-center gap-2 rounded-lg bg-primary-600 px-3 py-2 text-sm font-medium text-white hover:bg-primary-700 disabled:opacity-50"
        >
          {loading ? <Loader2 size={14} className="animate-spin" /> : <Sparkles size={14} />}
          {tab === 'suggest' && (loading ? 'Thinking…' : 'Suggest reply')}
          {tab === 'summary' && (loading ? 'Summarizing…' : 'Summarize thread')}
          {tab === 'sentiment' && (loading ? 'Analyzing…' : 'Analyze sentiment')}
        </button>

        {error && (
          <div className="rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
            {error}
          </div>
        )}

        {tab === 'suggest' && suggestion && (
          <div className="rounded-lg bg-gray-50 dark:bg-gray-700/40 p-3">
            <p className="text-xs text-gray-500 dark:text-gray-400 mb-2 uppercase tracking-wider">Suggested reply</p>
            <p className="text-sm text-gray-900 dark:text-gray-100 whitespace-pre-wrap">{suggestion}</p>
            <div className="flex gap-2 mt-3">
              <button
                onClick={() => onInsertReply(suggestion)}
                className="flex-1 rounded-lg bg-primary-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-primary-700"
              >
                Insert into reply
              </button>
              <button
                onClick={() => doCopy(suggestion)}
                className="rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-1.5 text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
                title="Copy"
              >
                {copied ? <Check size={12} /> : <Copy size={12} />}
              </button>
            </div>
          </div>
        )}

        {tab === 'summary' && summary && (
          <div className="rounded-lg bg-gray-50 dark:bg-gray-700/40 p-3">
            <p className="text-xs text-gray-500 dark:text-gray-400 mb-2 uppercase tracking-wider">Summary</p>
            <p className="text-sm text-gray-900 dark:text-gray-100 whitespace-pre-wrap">{summary}</p>
            <button
              onClick={() => doCopy(summary)}
              className="mt-3 flex items-center gap-1 text-xs text-primary-600 hover:text-primary-700"
            >
              {copied ? <Check size={12} /> : <Copy size={12} />} {copied ? 'Copied' : 'Copy'}
            </button>
          </div>
        )}

        {tab === 'sentiment' && sentiment && (
          <div className="rounded-lg bg-gray-50 dark:bg-gray-700/40 p-3 space-y-2">
            <p className="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wider">Sentiment</p>
            <div className="flex items-center gap-2">
              {sentiment.score === 'positive' && <><Smile size={20} className="text-green-600" /><span className="text-sm font-semibold text-green-700 dark:text-green-400 capitalize">{sentiment.score}</span></>}
              {sentiment.score === 'neutral' && <><Meh size={20} className="text-gray-500" /><span className="text-sm font-semibold text-gray-700 dark:text-gray-300 capitalize">{sentiment.score}</span></>}
              {sentiment.score === 'negative' && <><Frown size={20} className="text-red-600" /><span className="text-sm font-semibold text-red-700 dark:text-red-400 capitalize">{sentiment.score}</span></>}
            </div>
            <p className="text-xs text-gray-600 dark:text-gray-400 italic">{sentiment.reason}</p>
          </div>
        )}

        {!loading && !suggestion && !summary && !sentiment && !error && (
          <p className="text-xs text-gray-400 text-center py-6">
            Click the button above to run AI assist on this conversation.
          </p>
        )}
      </div>

      <div className="border-t border-gray-100 dark:border-gray-700 px-3 py-2 text-[10px] text-gray-400 text-center">
        Powered by Claude · claude.com
      </div>
    </aside>
  )
}

function TabBtn({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      className={`flex-1 flex items-center justify-center gap-1 py-2 font-medium transition-colors ${
        active
          ? 'text-primary-600 border-b-2 border-primary-600 -mb-px'
          : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
      }`}
    >
      {children}
    </button>
  )
}
