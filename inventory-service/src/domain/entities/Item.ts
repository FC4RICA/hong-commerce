export type Item = {
  id: string;
  name: string;
  quantity: number;
  reserved: number;
  available: number; // calculated field: quantity - reserved
  createdAt: Date;
  updatedAt: Date;
};

export const mapToItem = (data: any): Item => ({
  id: data.id,
  name: data.name,
  quantity: data.quantity,
  reserved: data.reserved,
  available: data.quantity - data.reserved,
  createdAt: data.createdAt,
  updatedAt: data.updatedAt,
});
