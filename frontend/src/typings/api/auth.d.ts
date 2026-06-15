declare namespace Api {
  /**
   * namespace Auth
   *
   * backend api module: "auth"
   */
  namespace Auth {
    interface LoginToken {
      token: string;
      refreshToken: string;
    }

    interface UserInfo {
      userId: string;
      userName: string;
      nickname: string;
      roles: string[];
      buttons: string[];
      permissions?: string[];
    }

    interface TwoFactorSetupInfo {
      qrCodeUrl: string;
      secret: string;
    }
  }
}
