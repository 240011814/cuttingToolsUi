declare namespace Api {
  namespace Admin {
    interface User {
      userId: number;
      userName: string;
      nickname: string;
      role: string;
      createdAt: string;
      updatedAt: string;
    }

    interface UserSearchParams {
      keyword?: string;
      role?: string;
    }

    interface CreateUserParams {
      userName: string;
      password: string;
      nickname: string;
      role: string;
    }

    interface UpdateUserParams {
      password?: string;
      nickname: string;
      role: string;
    }

    interface UserProfile {
      userId: number;
      userName: string;
      nickname: string;
      role: string;
      createdAt: string;
      updatedAt: string;
    }

    interface UpdateProfileParams {
      nickname: string;
    }

    interface ChangePasswordParams {
      oldPassword: string;
      newPassword: string;
    }

    interface Role {
      id: number;
      code: string;
      name: string;
      description: string;
    }

    interface Permission {
      id: number;
      code: string;
      name: string;
      groupName: string;
    }

    interface AIProvider {
      id: number;
      name: string;
      api_key: string;
      base_url: string;
      is_active: boolean;
      created_at: string;
      updated_at: string;
      models?: AIModel[];
    }

    interface AIModel {
      id: number;
      provider_id: number;
      model_code: string;
      display_name: string;
      is_default: boolean;
      config_json: string;
      created_at: string;
      updated_at: string;
    }
  }
}
