import { Button } from "../ui/button";
import { useNavigate } from "react-router-dom";

export function Unauthorized() {
  const navigate = useNavigate();
  
  return (
    <div className="flex flex-col items-center justify-center min-h-screen px-4 text-center">
      <h1 className="text-4xl font-bold mb-4">访问被拒绝</h1>
      <p className="text-muted-foreground mb-8">
        您没有权限访问此页面，请先登录
      </p>
      <Button onClick={() => navigate("/login")}>
        返回登录页
      </Button>
    </div>
  );
}