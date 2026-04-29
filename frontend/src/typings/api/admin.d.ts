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
  }
}
