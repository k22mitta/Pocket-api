import { consumeStream, convertToModelMessages, streamText, UIMessage } from 'ai'

export const maxDuration = 30

export async function POST(req: Request) {
  const {
    messages,
    financialContext,
  }: { messages: UIMessage[]; financialContext?: string } = await req.json()

  const systemPrompt = `You are Pocket, a sharp and concise personal finance assistant built into the Pocket app.
You help users understand their spending, budgets, and accounts.

Speak in plain, direct sentences — no jargon, no bullet-list overload. Keep answers focused and under 120 words unless a table or breakdown genuinely adds value.

${financialContext ? `Here is the user's current financial snapshot:\n\n${financialContext}` : 'No financial data has been loaded yet — let the user know they can connect accounts to get started.'}

Rules:
- Only discuss finances. Politely decline off-topic requests.
- Never make up transaction data or amounts not present in the snapshot.
- Format currency with a dollar sign and two decimal places (e.g. $1,234.56).
- Today is July 4, 2026.`

  const result = streamText({
    model: 'openai/gpt-4.1',
    system: systemPrompt,
    messages: await convertToModelMessages(messages),
    abortSignal: req.signal,
  })

  return result.toUIMessageStreamResponse({
    consumeSseStream: consumeStream,
  })
}
