import { NonDeletedExcalidrawElement, Ordered } from "@excalidraw/excalidraw/element/types";
import axios, { AxiosRequestConfig } from "axios";

export interface Drawing {
	readonly title: string;
	readonly content: string;
}

export type XcalidrawContent = Ordered<NonDeletedExcalidrawElement>[];

export type XcalidrawDocument = Readonly<{
	type: string;
	title: string;
	elements: XcalidrawContent;
}>;

const ITEM_TO_BE_DISCARDED = "___?????___";

export const fetchDrawingList = async (): Promise<string[]> => {
	const response = await axios.get("/api/drawings");
	return response.data.map((str: string) => base64ToPlain(str)).filter((str: string) => str !== ITEM_TO_BE_DISCARDED);
};

export const fetchDrawing = async (title: string) => {
	return await axios.get(`/api/drawing/${plainToBase64(title)}`);
};

export const saveDrawing = async (title: string, content: string) => {
	const doc: XcalidrawDocument = {
		type: "excalidraw",
		title,
		elements: JSON.parse(content)
	};
	await axios.put(`/api/drawing/${plainToBase64(title)}`, JSON.stringify({ content: JSON.stringify(doc) }), { headers: { "Content-Type": "application/json" } });
};

export const renameDrawing = async (from: string, to: string): Promise<void> => {
	console.log(">>>>>> renameDrawing: from: ", from, ", to:", to);
	const requestConfig: AxiosRequestConfig = {
		url: `/api/drawing/${plainToBase64(from)}`,
		method: "patch",
		data: JSON.stringify({
			newTitle: plainToBase64(to)
		}),
		headers: { "Content-Type": "application/json" }
	};
	await axios.request(requestConfig);
};

export const deleteDrawings = async (titles: string[]): Promise<void> => {
	console.log(">>>>>>>> deleteDrawings: ", titles.join(", "));
	for (const title of titles) {
		console.log(">>>>>>>> deleteDrawing ", title);
		await axios.delete(`/api/drawing/${plainToBase64(title)}`);
	}
};

export const base64ToPlain = (base64: string): string => {
	return new TextDecoder().decode(base64ToBytes(base64));
};

export const plainToBase64 = (plain: string): string => {
	return bytesToBase64(new TextEncoder().encode(plain));
};

const base64ToBytes = (base64: string): Uint8Array<ArrayBufferLike> => {
	try {
		const binString = atob(base64);
		return Uint8Array.from(binString, (m) => m.charCodeAt(0));
	} catch (err) {
		if ((err as DOMException).name === "InvalidCharacterError") {
			console.warn("Item with non-base64-encoded title is going to be discarded: ", base64);
			return new TextEncoder().encode(ITEM_TO_BE_DISCARDED);
		}
		throw err;
	}
};

const bytesToBase64 = (bytes: Uint8Array<ArrayBufferLike>): string => {
	const binString = Array.from(bytes, byte =>
		String.fromCodePoint(byte)
	).join("");
	return btoa(binString);
};
