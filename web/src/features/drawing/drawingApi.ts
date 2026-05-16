import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { ExcalidrawElement } from "@excalidraw/excalidraw/element/types";

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

export interface QualifiedDrawingId {
	readonly repoId: string;
	readonly drawingId: string;
}

export const qualifiedDrawingIdToString = (qualifiedDrawingId: QualifiedDrawingId): string =>
	`${qualifiedDrawingId.repoId}/${qualifiedDrawingId.drawingId}`;

export interface Drawing extends XcalidrawDocument {
	readonly id?: string;
	readonly repo: DrawingRepoRef;
}

export interface DrawingRepoItem {
	readonly id: string;
	readonly title: string;
}

export type DrawingRepo = Readonly<{
	repoRef: DrawingRepoRef;
	items: DrawingRepoItem[];
}>;

export type DrawingLists = Record<string, DrawingRepo>;

const TAG_DRAWING_LIST = "DrawingList";
const TAG_DRAWING = "Drawing";

export const drawingApi = createApi({
	reducerPath: "drawingApi",
	baseQuery: fetchBaseQuery({ baseUrl: "/api/" }),
	tagTypes: [TAG_DRAWING_LIST, TAG_DRAWING],
	endpoints: builder => ({
		getDrawingRepositories: builder.query<DrawingRepoRef[], void>({
			query: () => "drawingRepositories"
		}),
		getDrawingLists: builder.query<DrawingLists, void>({
			query: () => "drawings",
			providesTags: [TAG_DRAWING_LIST]
		}),
		getDrawing: builder.query<XcalidrawDocument, QualifiedDrawingId>({
			query: qualifiedDrawingId => `drawing/${qualifiedDrawingIdToString(qualifiedDrawingId)}`,
			transformResponse: (raw: string): XcalidrawDocument => JSON.parse(raw),
			providesTags: (_result, _error, arg) => [
				{ type: TAG_DRAWING, id: qualifiedDrawingIdToString(arg) }
			]
		}),
		createDrawing: builder.mutation<string, Drawing>({
			query: drawing => ({
				url: `drawing/${drawing.repo.name}`,
				method: "POST",
				body: {
					content: JSON.stringify(
						{ type: "excalidraw", title: drawing.title, elements: drawing.elements },
						null,
						"\t"
					)
				}
			}),
			invalidatesTags: [TAG_DRAWING_LIST]
		}),
		saveDrawing: builder.mutation<string, Drawing>({
			query: drawing => ({
				url: `drawing/${qualifiedDrawingIdToString({ repoId: drawing.repo.name, drawingId: drawing.id! })}`,
				method: "PUT",
				body: { content: JSON.stringify(drawing, null, "\t") }
			}),
			invalidatesTags: (_result, _error, arg) => [
				TAG_DRAWING_LIST,
				{
					type: TAG_DRAWING,
					id: qualifiedDrawingIdToString({ repoId: arg.repo.name, drawingId: arg.id! })
				}
			]
		}),
		deleteDrawings: builder.mutation<void, string[]>({
			queryFn: async (ids, _api, _extraOptions, baseQuery) => {
				for (const id of ids) {
					const result = await baseQuery({ url: `drawing/${id}`, method: "DELETE", responseHandler: "text" });
					if (result.error) {
						return { error: result.error };
					}
				}
				return { data: undefined };
			},
			invalidatesTags: [TAG_DRAWING_LIST]
		})
	})
});

export const {
	useGetDrawingRepositoriesQuery,
	useGetDrawingListsQuery,
	useLazyGetDrawingQuery,
	useCreateDrawingMutation,
	useSaveDrawingMutation,
	useDeleteDrawingsMutation
} = drawingApi;
