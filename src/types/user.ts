export interface GetUsersApi {
  users: User[];
}

export interface CreateUserApi {
  user: User;
}

export interface User {
  id: String;
  email: String;
  createdAt: String;
  updatedAt: String;
}
