import axios from "axios";

export interface Drawing {
  readonly title: string;
  readonly content: string;
}

export const fetchDrawingList = async () => {
  const response = await axios.get("/api/drawings");
  return response.data;
};

export const fetchDrawing = async (title: string) => {
  console.info("Fetching drawing: ", title);
  return await axios.get(`/api/drawing`, { params: { title } });
};

export const saveDrawing = async (title: string, content: string) => {
  console.info("Saving drawing of length ", content.length, ", as ", title);
  await axios.put(`/api/drawing`, JSON.stringify({ title, content }), { headers: { "Content-Type": "application/json" } });
};
