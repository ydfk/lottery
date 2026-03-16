import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar";
import { NavLink } from "react-router-dom";

const components = [
  {
    title: "按钮 Button",
    url: "/components/button",
  },
  {
    title: "输入框 Input",
    url: "/components/input",
  },
  {
    title: "多选框 Checkbox",
    url: "/components/checkbox",
  },
  {
    title: "单选组 RadioGroup",
    url: "/components/radio-group",
  },
  {
    title: "开关 Switch",
    url: "/components/switch",
  },
  {
    title: "滑块 Slider",
    url: "/components/slider",
  },
  {
    title: "选项卡 Tabs",
    url: "/components/tabs",
  },
  {
    title: "进度条 Progress",
    url: "/components/progress",
  },
  {
    title: "卡片 Card",
    url: "/components/card",
  },
  {
    title: "对话框 Dialog",
    url: "/components/dialog",
  },
  {
    title: "表单 Form",
    url: "/components/form",
  },
  {
    title: "表格 Table",
    url: "/components/table",
  },
  {
    title: "下拉菜单 Dropdown",
    url: "/components/dropdown",
  },
  {
    title: "提示 Sonner",
    url: "/components/sonner",
  },
  {
    title: "选择器 Select",
    url: "/components/select",
  },
  {
    title: "日历 Calendar",
    url: "/components/calendar",
  },
  {
    title: "Magic UI",
    url: "/components/magicui",
  },
];

export function AppSidebar() {
  return (
    <Sidebar collapsible="none" className="sticky top-0 h-svh border-r">
      <SidebarHeader>
        <div className="flex items-center px-4 py-2">
          <h2 className="text-lg font-semibold tracking-tight">Shadcn UI</h2>
        </div>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>综合案例</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton asChild>
                  <NavLink
                    to="/user-management"
                    className={({ isActive }) =>
                      isActive
                        ? "bg-sidebar-accent text-sidebar-accent-foreground"
                        : ""
                    }
                  >
                    <span>用户管理</span>
                  </NavLink>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        <SidebarGroup>
          <SidebarGroupLabel>组件示例</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {components.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild>
                    <NavLink
                      to={item.url}
                      className={({ isActive }) =>
                        isActive
                          ? "bg-sidebar-accent text-sidebar-accent-foreground"
                          : ""
                      }
                    >
                      <span>{item.title}</span>
                    </NavLink>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  );
}
