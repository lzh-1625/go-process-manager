import menuUI from "./menus/ui.menu";
import menuApps from "./menus/apps.menu";
import menuPages from "./menus/pages.menu";
import menuCharts from "./menus/charts.menu";
import menuUML from "./menus/uml.menu";
import menuLanding from "./menus/landing.menu";
import menuData from "./menus/data.menu";
import menuProcess from "./menus/process.menus";
import menuTask from "./menus/task.menus";
import menuUser from "./menus/user.menus";
import menuLog from "./menus/log.menus";
import menuSettings from "./menus/settings.menus";

export default {
  menu: [
    {
      text: "",
      key: "",
      items: [
        {
          key: "menu.dashboard",
          text: "Dashboard",
          link: "/dashboard",
          icon: "mdi-view-dashboard-outline",
        },
      ],
    },
    {
      text: "process",
      key: "menu.group.process",
      items: menuProcess,
    },
    {
      text: "task",
      key: "menu.group.task",
      items: menuTask,
    },
    {
      text: "log",
      key: "menu.group.log",
      items: menuLog,
    },
    {
      text: "user",
      key: "menu.group.user",
      items: menuUser,
    },
    {
      text: "settings",
      key: "menu.group.settings",
      items: menuSettings,
    },
    {
      text: "Apps",
      key: "menu.group.apps",
      items: menuApps,
    },
    {
      text: "Data",
      key: "menu.group.data",
      items: menuData,
    },
    {
      text: "Landing",
      key: "menu.group.landing",
      items: [
        ...menuLanding,
        // {
        //   icon: "mdi-airplane-landing",
        //   key: "menu.landingPage",
        //   text: "Landing Page",
        //   link: "/landing",
        // },
      ],
    },

    {
      text: "UI - Theme Preview",
      key: "menu.group.ui",
      items: menuUI,
    },
    {
      text: "Pages",
      key: "menu.group.pages",
      items: menuPages,
    },
    {
      text: "Charts",
      key: "menu.group.charts",
      items: menuCharts,
    },
    {
      text: "UML",
      key: "menu.group.uml",
      items: menuUML,
    },
  ],
};
