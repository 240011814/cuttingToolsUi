declare namespace Api {
  namespace Telegram {
    interface StatusResponse {
      isBound: boolean;
      telegramUsername?: string;
    }

    interface BindCodeResponse {
      bindCode: string;
      expiresAt: number;
      botName: string;
    }
  }
}
