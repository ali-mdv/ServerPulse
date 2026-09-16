import { createRouter, createWebHistory, RouteRecordRaw } from "vue-router";
import Login from "@/pages/Login.vue";
import Dashboard from "@/pages/Dashboard.vue";
import Alerts from "@/pages/Alerts.vue";
import Settings from "@/pages/Settings.vue";
import Servers from "@/pages/Servers.vue";
import ServerDetails from "@/pages/ServerDetails.vue";
import Users from "@/pages/Users.vue";
import AddUser from "@/pages/AddUser.vue";
import Profile from "@/pages/Profile.vue";
import ServerError from "@/pages/ServerError.vue";
import NotFound from "@/pages/NotFound.vue";
import { useAuthStore } from "@/stores/auth";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    redirect: "/dashboard",
    name: "home",
  },
  {
    path: "/auth",
    redirect: "/auth/login",
    children: [
      {
        path: "/login",
        component: Login,
        name: "login",
      },
    ],
  },
  {
    path: "/dashboard",
    component: Dashboard,
    name: "dashboard",
    meta: { requiresAuth: true },
  },
  {
    path: "/servers",
    component: Servers,
    name: "serversList",
    meta: { requiresAuth: true },
  },
  {
    path: "/server/:id",
    component: ServerDetails,
    name: "serversDetail",
    meta: { requiresAuth: true },
  },
  {
    path: "/alerts",
    component: Alerts,
    name: "alerts",
    meta: { requiresAuth: true },
  },
  {
    path: "/users",
    component: Users,
    name: "usersList",
    meta: { requiresAuth: true },
  },
  {
    path: "/users/add",
    component: AddUser,
    name: "usersAdd",
    meta: { requiresAuth: true },
  },
  {
    path: "/profile",
    component: Profile,
    name: "profile",
    meta: { requiresAuth: true },
  },
  {
    path: "/settings",
    component: Settings,
    name: "settings",
    meta: { requiresAuth: true },
  },
  {
    path: "/500",
    component: ServerError,
    name: "serverError",
  },
  {
    path: "/:pathMatch(.*)*",
    component: NotFound,
    name: "notFound",
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to) => {
  const auth = useAuthStore();
  auth.checkAuthentication();

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: "login" };
  }

  if (to.name === "login" && auth.isAuthenticated) {
    return { name: "dashboard" };
  }
});

export default router;
