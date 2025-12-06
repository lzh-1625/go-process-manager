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
      permission: 1,
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
      permission: 1,
    },
    {
      text: "log",
      key: "menu.group.log",
      items: menuLog,
      permission: 2,
    },
    {
      text: "user",
      key: "menu.group.user",
      items: menuUser,
      permission: 0,
    },
    {
      text: "settings",
      key: "menu.group.settings",
      items: menuSettings,
    },
  ],
};
