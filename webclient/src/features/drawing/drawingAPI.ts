import { ExcalidrawElement } from "@excalidraw/excalidraw/element/types";
import axios from "axios";

export type XcalidrawContent = ExcalidrawElement[];

export type XcalidrawDocument = Readonly<{
	title: string;
	elements: XcalidrawContent;
	type?: string;
}>;

export interface DrawingRepoRef {
	readonly name: string;
	readonly label: string;
}

export interface Drawing extends XcalidrawDocument {
	readonly id?: string;
	readonly repo?: DrawingRepoRef;
}

export interface DrawingRepoItem {
	readonly id: string;
	readonly title: string;
}

export type DrawingRepo = Readonly<{
	repoRef: DrawingRepoRef;
	items: DrawingRepoItem[];
}>;
export type DrawingLists = Record<string /*DrawingRepoName*/, DrawingRepo>;

export const fetchDrawingList = async (): Promise<DrawingLists> => {
	const response = await axios.get("/api/drawings");
	const drawingRepoNameToDrawingRepoContent: DrawingLists = response.data;
	return drawingRepoNameToDrawingRepoContent;
};

export const fetchDrawing = async (id: string): Promise<Drawing> => {
	const response = await axios.get(`/api/drawing/${id}`);
	return JSON.parse(response.data);
};

export const createDrawing = async (title: string, elements: XcalidrawContent): Promise<string> => {
	const doc: XcalidrawDocument = {
		type: "excalidraw",
		title,
		elements
	};
	const response = await axios.post("/api/drawing", JSON.stringify({ content: JSON.stringify(doc, null, "\t") }), { headers: { "Content-Type": "application/json" } });
	return response.data;
};

export const saveDrawing = async (drawing: Drawing): Promise<string> => {
	const response = await axios.put(`/api/drawing/${drawing.id}`, JSON.stringify({ content: JSON.stringify(drawing, null, "\t") }), { headers: { "Content-Type": "application/json" } });
	return response.data;
};

export const deleteDrawings = async (ids: string[]): Promise<void> => {
	console.log(">>>>>>>> deleteDrawings: ", ids.join(", "));
	for (const id of ids) {
		console.log(">>>>>>>> deleteDrawing ", id);
		await axios.delete(`/api/drawing/${id}`);
	}
};
