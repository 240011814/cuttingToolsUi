let voicesReady = false;

/**
 * 等待 Chrome voices 加载完成
 */
const ensureVoices = (): Promise<SpeechSynthesisVoice[]> => {
  return new Promise((resolve) => {
    const synth = window.speechSynthesis;

    const voices = synth.getVoices();
    if (voices.length > 0) {
      voicesReady = true;
      resolve(voices);
      return;
    }

    const handler = () => {
      const v = synth.getVoices();
      voicesReady = true;
      synth.removeEventListener('voiceschanged', handler);
      resolve(v);
    };

    synth.addEventListener('voiceschanged', handler);
  });
};

/**
 * 播放语音（稳定版）
 */
export const speak = async (
  text: string,
  options?: {
    lang?: string;
    rate?: number;
  }
) => {
  if (!window.speechSynthesis) {
    throw new Error('SpeechSynthesis not supported');
  }

  const synth = window.speechSynthesis;

  // 1. 确保 voices ready（Chrome 必须）
  if (!voicesReady) {
    await ensureVoices();
  }

  // 2. 停止之前播放
  synth.cancel();

  const utterance = new SpeechSynthesisUtterance(text);

  utterance.lang = options?.lang || 'en-US';
  utterance.rate = options?.rate ?? 0.9;

  // 3. 选择 voice（更稳定）
  const voices = synth.getVoices();
  const voice = voices.find(v => v.lang === utterance.lang);
  if (voice) {
    utterance.voice = voice;
  }

  // 4. 返回 Promise，播放结束时 resolve
  return new Promise<void>((resolve, reject) => {
    utterance.addEventListener('end', () => resolve());
    utterance.addEventListener('error', (e) => {
      console.error('TTS error:', e);
      reject(e);
    });

    // Chrome 稳定技巧：延迟一帧
    requestAnimationFrame(() => {
      synth.speak(utterance);
    });
  });
};

/**
 * 停止播放
 */
export const stopSpeak = () => {
  if (window.speechSynthesis) {
    window.speechSynthesis.cancel();
  }
};
