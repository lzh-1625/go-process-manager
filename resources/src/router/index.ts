import { createRouter, createWebHistory } from "vue-router";
import UserRoutes from "./user.routes";
import ProcessRoutes from "./process.routes";
import TaskRoutes from "./task.routes";
import LogRoutes from "./log.routes";
import SettingsRoutes from "./settings.routes";

export const routes = [
  {
    path: "/",
    redirect: parseInt(window.localStorage.getItem("role") ?? "0") < 2 ? "/dashboard" : "/process",
    meta: {},
  } as any,
  {
    path: "/dashboard",
    meta: {
      requiresAuth: true,
      layout: "landing",
    },
    component: () => import("@/views/pages/DashBoard.vue"),
  },
  {
    path: "/login",
    component: () => import("@/views/login/Login.vue"),
    meta: {
      requiresAuth: false,
      layout: "auth",
    },
  },
  {
    path: "/share",
    component: () => import("@/views/share/ShareTerminal.vue"),
    meta: {
      requiresAuth: false,
      layout: "blank",
    },
  },
  {
    path: "/:pathMatch(.*)*",
    name: "error",
    component: () =>
      import(/* webpackChunkName: "error" */ "@/views/errors/NotFoundPage.vue"),
  },
  ...UserRoutes,
  ...ProcessRoutes,
  ...TaskRoutes,
  ...LogRoutes,
  ...SettingsRoutes
];

// 动态路由，基于用户权限动态去加载
export const dynamicRoutes = [];

const router = createRouter({
  history: createWebHistory(),
  // hash模式：createWebHashHistory，history模式：createWebHistory
  // process.env.NODE_ENV === "production"

  routes: routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition;
    } else {
      return { top: 0 };
    }
  },
});

export default router;
