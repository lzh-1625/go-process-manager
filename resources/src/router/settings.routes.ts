// Settings Page Routes
export default [
  {
    path: "/settings",
    component: () => import("@/views/settings/Settings.vue"),
    meta: {
      requiresAuth: true,
      layout: "landing",
      category: "Settings",
    },
  },
  {
    path: "/settings/push",
    component: () => import("@/views/settings/Push.vue"),
    meta: {
      requiresAuth: true,
      layout: "landing",
      category: "Settings",
    },
  },
];

