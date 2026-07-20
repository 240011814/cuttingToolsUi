declare namespace Api {
  namespace Telegram {
    interface ConfigResponse {
      configured: boolean;
    }

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
