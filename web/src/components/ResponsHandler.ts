
export function handleError(error: any, setError: (message: string | null) => void) {
  const isHttpError = error instanceof Error && (error.message || "").includes("HTTP error");

  if (isHttpError) {
      const genericServerError = "An unexpected error occurred on the server.";

      if (error.message.includes("400")) {
        setError("Invalid input. Please check the data and try again.");
      } else if (error.message.includes("500")) {
        setError("A server error occurred. Please try again later.");
      }

      else if (error.message.includes("409")) {
        if (error.message.includes("Path already exists")) {
            console.error("Conflict: Path already exists.");
            setError("The provided path already exists.");
        } else if (error.message.includes("Admin user already exists")) {
            console.error("Conflict: Admin user already exists.");
            setError("The provided admin user already exists.");
        }  else {
            console.error("Conflict: An unidentified conflict occurred.");
            setError(genericServerError);
        }
      } else {
          console.error("Unhandled HTTP error:", error);
          setError(genericServerError);
      }

      const timer = setTimeout(() => {
          setError(null);
      }, 4000);
      return () => clearTimeout(timer); 
  }

  // Generic error for Non-HTTP error
  console.error("Non-HTTP error occurred:", error);
  setError("An unexpected error occurred. Please try again.");
  const timer = setTimeout(() => {
      setError(null);
  }, 4000);
  return () => clearTimeout(timer);
}