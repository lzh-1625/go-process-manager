// Log Page Routes
export default [
  {
    path: "/log",
    component: () => import("@/views/log/Log.vue"),
    meta: {
      requiresAuth: true,
      layout: "landing",
      category: "Data",
    },
  },
];

