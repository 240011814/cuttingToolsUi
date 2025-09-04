declare namespace Api {
  /**
   * namespace Auth
   *
   * backend api module: "auth"
   */
  namespace Auth {
    interface LoginToken {
      accessToken: string;
      refreshToken: string;
    }

    interface UserInfo {
      userId: string;
      userName: string;
      roles: string[];
      buttons: string[];
    }

    interface RegisterRequest {
      userName: string;
      phone: string;
      code: string;
      password: string;
      confirmPassword: string;
    }
  }
}
